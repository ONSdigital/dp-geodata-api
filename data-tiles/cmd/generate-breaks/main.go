package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/cat"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/geos"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/grid"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/types"
	"github.com/jtrim-ons/ckmeans/pkg/ckmeans"
)

func main() {
	catfile := flag.String("c", "categories.txt", "text file holding list of categories to use")
	gridfile := flag.String("q", "DataTileGrid.json", "json tile description file")
	geodir := flag.String("G", "data/processed/geo", "directory holding geojson files for each geotype")
	metdir := flag.String("M", "data/processed/metrics", "directory holding metrics files for each category")
	outdir := flag.String("O", "data/output/breaks", "output directory")
	calcRatios := flag.Bool("R", false, "calculate ratios")
	flag.Parse()

	quads, err := grid.Load(*gridfile)
	if err != nil {
		log.Fatal(err)
	}
	var wanttypes []types.Geotype
	for geotype := range quads {
		wanttypes = append(wanttypes, geotype)
	}

	geotypes, err := loadGeotypes(*geodir)
	if err != nil {
		log.Fatal(err)
	}

	catlist, err := cat.LoadCategories(*catfile)
	if err != nil {
		log.Fatal(err)
	}

	// if we are calculating ratios, we also want to load the
	// totals categories
	var loadcats []types.Category
	if !*calcRatios {
		loadcats = catlist
	} else {
		allcats, err := cat.IncludeTotalCats(catlist)
		if err != nil {
			log.Fatal(err)
		}
		loadcats = allcats
	}

	metrics, err := cat.LoadMetrics(loadcats, *metdir)
	if err != nil {
		log.Fatal(err)
	}

	if err := genbreaks(wanttypes, geotypes, catlist, metrics, *calcRatios, *outdir); err != nil {
		log.Fatal(err)
	}
}

// loadGeotypes read geojson files and build a map from geocode to geotype
func loadGeotypes(dir string) (map[types.Geocode]types.Geotype, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.geojson"))
	if err != nil {
		return nil, err
	}

	geotypes := make(map[types.Geocode]types.Geotype)

	for _, fname := range matches {
		log.Printf("Loading %s\n", fname)

		geojson, err := geos.LoadGeojson(fname)
		if err != nil {
			return nil, err
		}

		for i, feat := range geojson.Features {
			typeprop := feat.Properties["geotype"].(string)
			codeprop := feat.Properties["geocode"].(string)
			if typeprop == "" || codeprop == "" {
				log.Printf("\tfeature %d: missing geotype or geoname", i)
				continue
			}
			geotypes[types.Geocode(codeprop)] = types.Geotype(typeprop)
		}
	}
	return geotypes, nil
}

type stats map[types.Category]map[string][]float64

// genbreaks generates the breaks files
func genbreaks(
	wanttypes []types.Geotype,
	geotypes map[types.Geocode]types.Geotype,
	cats []types.Category,
	metrics map[types.Category]map[types.Geocode]types.Value,
	calcMetrics bool,
	dir string,
) error {
	if calcMetrics {
		return genbreaksWithRatios(wanttypes, geotypes, cats, metrics, dir)
	} else {
		return genbreaksWithoutRatios(wanttypes, geotypes, cats, metrics, dir)
	}
}

func genbreaksWithRatios(
	wanttypes []types.Geotype,
	geotypes map[types.Geocode]types.Geotype,
	cats []types.Category,
	metrics map[types.Category]map[types.Geocode]types.Value,
	dir string,
) error {
	for _, thiscat := range cats {
		fmt.Fprintf(os.Stderr, " %s\n", thiscat)

		totcat, err := cat.GuessTotalsCat(thiscat)
		if err != nil {
			return err
		}

		ratios := make(map[types.Geotype][]float64)

		for geocode, value := range metrics[thiscat] {
			geotype, ok := geotypes[geocode]
			if !ok {
				log.Printf(
					"%s: no geotype for %s, skipping geo",
					thiscat,
					geocode,
				)
				continue
			}
			total, ok := metrics[totcat][geocode]
			if !ok {
				return fmt.Errorf(
					"no %s %s %s value",
					geotype,
					totcat,
					geocode,
				)
			}

			if total == 0.0 {
				return fmt.Errorf(
					"%s %s %s is 0",
					geotype,
					totcat,
					geocode,
				)
			}

			ratio := float64(value / total)

			ratios[geotype] = append(ratios[geotype], ratio)
		}

		for geotype, values := range ratios {
			breaks, err := getBreaks(values, 5)
			if err != nil {
				return fmt.Errorf(
					"%s %s (%d values): %w",
					geotype,
					thiscat,
					len(values),
					err,
				)
			}

			minmax := getMinMax(values)

			result := stats{
				thiscat: map[string][]float64{
					string(geotype):              breaks,
					string(geotype) + "_min_max": minmax,
				},
			}

			if err := saveStats(dir, geotype, thiscat, result); err != nil {
				return fmt.Errorf("%s %s: %w", geotype, thiscat, err)
			}
		}
	}
	return nil
}

func genbreaksWithoutRatios(
	wanttypes []types.Geotype,
	geotypes map[types.Geocode]types.Geotype,
	cats []types.Category,
	metrics map[types.Category]map[types.Geocode]types.Value,
	dir string,
) error {
	for _, thiscat := range cats {
		fmt.Fprintf(os.Stderr, " %s\n", thiscat)

		ratios := make(map[types.Geotype][]float64)

		for geocode, value := range metrics[thiscat] {
			geotype, ok := geotypes[geocode]
			if !ok {
				log.Printf(
					"%s: no geotype for %s, skipping geo",
					thiscat,
					geocode,
				)
				continue
			}
			ratio := value
			ratios[geotype] = append(ratios[geotype], float64(ratio))
		}

		for geotype, values := range ratios {
			breaks, err := getBreaks(values, 5)
			if err != nil {
				return fmt.Errorf(
					"%s %s (%d values): %w",
					geotype,
					thiscat,
					len(values),
					err,
				)
			}

			minmax := getMinMax(values)

			result := stats{
				thiscat: map[string][]float64{
					string(geotype):              breaks,
					string(geotype) + "_min_max": minmax,
				},
			}

			if err := saveStats(dir, geotype, thiscat, result); err != nil {
				return fmt.Errorf("%s %s: %w", geotype, thiscat, err)
			}
		}
	}
	return nil
}

// getBreaks gets k ckmeans clusters from metrics and returns the upper breakpoints for each cluster.
func getBreaks(metrics []float64, k int) ([]float64, error) {
	clusters, err := ckmeans.Ckmeans(metrics, k)
	if err != nil {
		return nil, err
	}

	var breaks []float64
	for _, cluster := range clusters {
		bp := cluster[len(cluster)-1]
		breaks = append(breaks, bp)
	}
	return breaks, nil
}

func getMinMax(values []float64) []float64 {
	max := values[0]
	min := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	return []float64{min, max}
}

func saveStats(dir string, geotype types.Geotype, cat types.Category, result stats) error {
	d := filepath.Join(dir, geotype.Pathname())
	if err := os.MkdirAll(d, 0755); err != nil {
		return err
	}

	name := filepath.Join(d, string(cat)+".json")

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(name, data, 0644)
}
