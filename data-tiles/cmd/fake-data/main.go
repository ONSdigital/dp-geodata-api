package main

import (
	"encoding/csv"
	"flag"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/content"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/geos"
)

func main() {
	// load content.json and get categories
	cfile := flag.String("c", "cmd/fake-data/testdata/small-content.json", "path to content.json")
	geodir := flag.String("G", "cmd/fake-data/testdata", "directory holding geojson files ")
	seed := flag.Int64("r", 0, "random number seed")
	flag.Parse()

	log.Printf("Loading %s\n", *cfile)
	c, err := content.LoadName(*cfile)
	if err != nil {
		log.Fatal(err)
	}

	cats := getCats(c)
	log.Printf("\tfound %d categories\n", len(cats))

	geos, err := loadGeos(*geodir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("\tfound %d geocodes\n", len(geos))

	if err := genFake(cats, geos, *seed); err != nil {
		log.Fatal(err)
	}
}

func getCats(c *content.Content) []string {
	var cats []string
	for _, group := range c.TopicGroups {
		for _, topic := range group.Topics {
			for _, variable := range topic.Variables {
				for _, classification := range variable.Classifications {
					for _, cat := range classification.Categories {
						cats = append(cats, cat.Code)
					}
				}
			}
		}
	}
	return cats
}

func loadGeos(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.geojson"))
	if err != nil {
		return nil, err
	}

	var geolist []string
	for _, fname := range matches {
		log.Printf("Loading %s\n", fname)
		geojson, err := geos.LoadGeojson(fname)
		if err != nil {
			return nil, err
		}

		for _, feat := range geojson.Features {
			geo, ok := feat.Properties["geocode"].(string)
			if !ok || geo == "" {
				continue
			}
			geolist = append(geolist, geo)
		}
	}
	return geolist, nil
}

func genFake(cats, geos []string, seed int64) error {
	rnd := rand.New(rand.NewSource(seed))

	w := csv.NewWriter(os.Stdout)

	headings := append([]string{"GeographyCode"}, cats...)
	if err := w.Write(headings); err != nil {
		return err
	}

	for _, geo := range geos {
		row := []string{geo}
		for i := 0; i < len(cats); i++ {
			val := rnd.Float64()
			cell := strconv.FormatFloat(val, 'f', 6, 64)
			row = append(row, cell)
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
