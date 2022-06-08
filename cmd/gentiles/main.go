package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ONSdigital/dp-geodata-api/pkg/database"
	"github.com/ONSdigital/dp-geodata-api/pkg/geodata"
	dplog "github.com/ONSdigital/log.go/v2/log"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const (
	totalsuffix = "0001"
	breaksdir   = "breaks"
)

type tile struct {
	Name string `json:"tilename"`
	Bbox box    `json:"bbox"`
}

type box struct {
	East  float64 `json:"east"`
	North float64 `json:"north"`
	West  float64 `json:"west"`
	South float64 `json:"south"`
}

type generator struct {
	dir      string            // base directory for output files
	tiles    map[string][]tile // tiles defined in spec file
	totfiles int               // total number of files we expect to create
	db       *database.Database
	app      *geodata.Geodata
	cats     []string // list of categories we create files and ckmeans for
	geos     []string // geocodes to create files for
	start    time.Time
	ch       chan struct{}
	wg       sync.WaitGroup
	nfiles   int32
}

func main() {
	dplog.SetDestination(os.Stderr, os.Stdout)
	dir := flag.String("o", ".", "output directory")
	concurrency := flag.Int("j", 1, "concurrency")
	catfile := flag.String("C", "", "name of file holding categories, one per line")
	tilefile := flag.String("T", "", "name of file holding tile names and bboxes")
	geofile := flag.String("G", "", "name of file holding geocodes")
	flag.Parse()
	if *dir == "" {
		log.Fatal("must supply output directory(-o)")
	}

	if err := os.Mkdir(*dir, 0755); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	ctx := context.Background()

	var totfiles int

	var tiles map[string][]tile
	var cats []string
	if *tilefile != "" && *catfile != "" {
		var err error
		tiles, err = loadtiles(*tilefile)
		if err != nil {
			log.Fatal(err)
		}

		for _, tiles := range tiles {
			totfiles += len(tiles)
		}
		cats, err = loadlist(*catfile)
		if err != nil {
			log.Fatal(err)
		}
		totfiles += len(tiles) * len(cats)
	}

	var geos []string
	if *geofile != "" {
		var err error
		geos, err = loadlist(*geofile)
		if err != nil {
			log.Fatal(err)
		}
		totfiles += len(geos)
	}

	db, err := database.Open("pgx", database.GetDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app, err := geodata.New(db, nil, 400000)
	if err != nil {
		log.Fatal(err)
	}

	generator := &generator{
		dir:      *dir,
		tiles:    tiles,
		totfiles: totfiles,
		db:       db,
		app:      app,
		cats:     cats,
		geos:     geos,
		start:    time.Now(),
		ch:       make(chan struct{}, *concurrency),
		nfiles:   0,
	}

	if *tilefile != "" && *catfile != "" {
		if err = generator.gentiles(ctx); err != nil {
			log.Fatal(err)
		}

		if err = generator.genckmeans(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if *geofile != "" {
		if err = generator.gengeos(ctx); err != nil {
			log.Fatal(err)
		}
	}

	generator.wg.Wait()
	fmt.Printf("total elapsed time: %s\n", time.Since(generator.start))
}

func loadtiles(name string) (map[string][]tile, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	specs := map[string][]tile{}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&specs); err != nil {
		return nil, err
	}

	return specs, nil
}

func (g *generator) gentiles(ctx context.Context) error {
	for geotype, tiles := range g.tiles {
		if err := os.Mkdir(filepath.Join(g.dir, geotype), 0755); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}

		for _, tile := range tiles {
			if err := os.Mkdir(filepath.Join(g.dir, geotype, tile.Name), 0755); err != nil && !os.IsExist(err) {
				log.Fatal(err)
			}

			for _, category := range g.cats {
				g.ch <- struct{}{}
				g.wg.Add(1)
				n := atomic.AddInt32(&g.nfiles, 1)
				go func(n int32, geotype, tilename, cat string, bbox box) {
					defer func() {
						g.wg.Done()
						<-g.ch
					}()

					prefix, suffix := splitcat(cat)
					if suffix == totalsuffix {
						fmt.Printf("%d/%d %s (skipping totals category)\n", n, g.totfiles, cat)
						return
					}

					var err error
					fn := filepath.Join(g.dir, geotype, tilename, cat+".csv")
					if already, err := exists(fn); err != nil {
						log.Fatal(err)
					} else if already {
						fmt.Printf("%d/%d %s (exists)\n", n, g.totfiles, fn)
						return
					} else {
						g.printstatus(fn, int(n))
					}

					body, err := g.app.Query(
						ctx,
						2011,
						fmt.Sprintf("%.13g,%.13g,%.13g,%.13g", bbox.East, bbox.South, bbox.West, bbox.North),
						"",
						0,
						"",
						[]string{geotype},
						nil,
						[]string{"geography_code", cat},
						"",
						prefix+totalsuffix,
					)
					if err != nil {
						log.Fatal(err)
					}

					if err = os.WriteFile(fn+".tmp", []byte(body), 0644); err != nil {
						log.Fatal(err)
					}
					if err = os.Rename(fn+".tmp", fn); err != nil {
						log.Fatal(err)
					}
				}(n, geotype, tile.Name, category, tile.Bbox)
			}
		}
	}
	return nil
}

func (g *generator) genckmeans(ctx context.Context) error {
	if err := os.Mkdir(filepath.Join(g.dir, breaksdir), 0755); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	for geotype, _ := range g.tiles {
		if err := os.Mkdir(filepath.Join(g.dir, breaksdir, geotype), 0755); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}

		for _, cat := range g.cats {
			g.ch <- struct{}{}
			g.wg.Add(1)
			n := atomic.AddInt32(&g.nfiles, 1)
			go func(geotype, cat string) {
				defer func() {
					g.wg.Done()
					<-g.ch
				}()

				prefix, suffix := splitcat(cat)
				if suffix == totalsuffix {
					return
				}

				var err error
				fn := filepath.Join(g.dir, breaksdir, geotype, cat+".json")
				if already, err := exists(fn); err != nil {
					log.Fatal(err)
				} else if already {
					fmt.Printf("%s (exists)\n", fn)
				} else {
					g.printstatus(fn, int(n))
				}

				breaks, err := g.app.CKmeans(ctx, 2011, []string{cat}, []string{geotype}, 5, prefix+totalsuffix)
				if err != nil {
					log.Fatal(err)
				}

				buf, err := json.MarshalIndent(breaks, "", "    ")
				if err != nil {
					log.Fatal(err)
				}

				if err = os.WriteFile(fn+".tmp", buf, 0644); err != nil {
					log.Fatal(err)
				}
				if err = os.Rename(fn+".tmp", fn); err != nil {
					log.Fatal(err)
				}
			}(geotype, cat)
		}
	}
	return nil
}

func (g *generator) gengeos(ctx context.Context) error {
	if err := os.Mkdir(filepath.Join(g.dir, "geo"), 0775); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	for _, geocode := range g.geos {
		g.ch <- struct{}{}
		g.wg.Add(1)
		n := atomic.AddInt32(&g.nfiles, 1)
		go func(geocode string) {
			defer func() {
				g.wg.Done()
				<-g.ch
			}()

			var err error
			fn := filepath.Join(g.dir, "geo", geocode+".geojson")
			if already, err := exists(fn); err != nil {
				log.Fatal(err)
			} else if already {
				fmt.Printf("%s (exists)\n", fn)
			} else {
				g.printstatus(fn, int(n))
			}

			resp, err := g.app.Geo(ctx, 2011, geocode, "")
			if err != nil {
				log.Fatal(err)
			}
			//resp.GeoJSON.Features[1] = resp.GeoJSON.Features[2]
			//resp.GeoJSON.Features = resp.GeoJSON.Features[0:2]

			if resp.GeoJSON == nil {
				return // some geographies do not have geometries
			}

			// squeeze out "boundary" feature
			var features []*geojson.Feature
			for _, feat := range resp.GeoJSON.Features {
				if feat.ID == "centroid" || feat.ID == "bbox" {
					features = append(features, feat)
				}
			}
			resp.GeoJSON.Features = features

			content, err := json.MarshalIndent(resp, "", "    ")
			if err != nil {
				log.Fatal(err)
			}
			content = append(content, "\n"...)

			if err = os.WriteFile(fn+".tmp", content, 0644); err != nil {
				log.Fatal(err)
			}
			if err = os.Rename(fn+".tmp", fn); err != nil {
				log.Fatal(err)
			}
		}(geocode)
	}
	return nil
}

func (g *generator) printstatus(fn string, n int) {
	est := status(g.start, n, g.totfiles)
	fmt.Printf(
		"%d/%d (%0.2f%%) %s [%s/%s] finish=%s\n",
		n,
		g.totfiles,
		est.PctDone,
		fn,
		est.Remain.Round(time.Second),
		est.Duration.Round(time.Second),
		est.Finish.Truncate(time.Second),
	)
}

func loadlist(fname string) ([]string, error) {
	buf, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return listok(strings.Split(string(buf), "\n"))
}

func listok(items []string) ([]string, error) {
	var newitems []string
	for _, item := range items {
		if item != "" {
			newitems = append(newitems, item)
		}
	}
	if len(newitems) == 0 {
		return nil, errors.New("empty list")
	}
	return newitems, nil
}

func splitcat(cat string) (left, right string) {
	return cat[0 : len(cat)-4], cat[len(cat)-4:]
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type estimate struct {
	PctDone  float64
	Duration time.Duration // estimated total duration of run
	Remain   time.Duration // estimated time remaining
	Finish   time.Time     // estimated finish time
}

func status(start time.Time, completed, total int) estimate {
	elapsed := time.Since(start)
	duration := time.Duration((float64(elapsed) * float64(total)) / float64(completed))
	remain := duration - elapsed
	finish := start.Add(duration)

	return estimate{
		PctDone:  (float64(completed) / float64(total)) * 100,
		Duration: duration,
		Remain:   remain,
		Finish:   finish,
	}
}
