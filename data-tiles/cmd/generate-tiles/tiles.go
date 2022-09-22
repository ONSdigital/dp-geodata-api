package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/cat"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/grid"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/types"

	"github.com/twpayne/go-geom"
)

func generateTiles(
	cats []types.Category,
	quadset map[types.Geotype][]grid.Quad,
	bounds map[types.Geotype]map[types.Geocode]*geom.Bounds,
	metrics map[types.Category]map[types.Geocode]types.Value,
	calcRatios bool,
	dir string,
) error {
	// for every geotype in DataTilesGrid.json
	for geotype, quadlist := range quadset {
		fmt.Printf("Generating tiles for %s\n", geotype)

		// for every quad of this geotype
		for _, q := range quadlist {
			// find overlapping geographies of this geotype from geojson
			geos := findOverlaps(q.Bbox, bounds[geotype])
			log.Printf(
				"Selected %d %s geographies for tile %s",
				len(geos),
				geotype,
				q.Tilename,
			)

			// generate category files for this quad
			err := generateCatFiles(
				geotype,
				q.Tilename,
				cats,
				geos,
				metrics,
				calcRatios,
				dir,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// findOverlaps selects geographies in 'bounds' that overlap bbox.
func findOverlaps(bbox *geom.Bounds, bounds map[types.Geocode]*geom.Bounds) []types.Geocode {
	var overlaps []types.Geocode
	for geocode, geobounds := range bounds {
		if bbox.Overlaps(geom.XY, geobounds) {
			overlaps = append(overlaps, geocode)
		}
	}
	return overlaps
}

func generateCatFiles(
	geotype types.Geotype,
	tilename string,
	cats []types.Category,
	geos []types.Geocode,
	metrics map[types.Category]map[types.Geocode]types.Value,
	wantRatios bool,
	dir string,
) error {
	d := filepath.Join(dir, geotype.Pathname(), tilename)
	if err := os.MkdirAll(d, 0755); err != nil {
		return err
	}

	for _, thiscat := range cats {
		// extract metrics for selected geos
		values, err := extractMetrics(geos, metrics[thiscat])
		if err != nil {
			return fmt.Errorf("%s %s: %w", geotype, thiscat, err)
		}

		if wantRatios {
			totcat, err := cat.GuessTotalsCat(thiscat)
			if err != nil {
				return fmt.Errorf("%s %s: %w", geotype, thiscat, err)
			}

			ratios, err := calcRatios(values, metrics[totcat])
			if err != nil {
				return fmt.Errorf("%s %s: %w", geotype, thiscat, err)
			}
			values = ratios
		}

		// save ratios to tile's category file
		if err := writeCatFile(d, thiscat, values); err != nil {
			return err
		}
	}
	return nil
}

// extractMetrics returns metrics only for selected geos
func extractMetrics(geos []types.Geocode, metrics map[types.Geocode]types.Value) (map[types.Geocode]types.Value, error) {
	result := map[types.Geocode]types.Value{}
	for _, geo := range geos {
		// The geojson may contain geocodes from areas out of scope
		// and we may not have metrics for these areas.
		value, ok := metrics[geo]
		if ok {
			result[geo] = value
		}
	}
	return result, nil
}

// calcRatios calculates the ratio of each value over the total
func calcRatios(values, totals map[types.Geocode]types.Value) (map[types.Geocode]types.Value, error) {
	result := map[types.Geocode]types.Value{}

	for geocode, value := range values {
		tot, exists := totals[geocode]
		if !exists {
			return nil, fmt.Errorf("geocode %s: no totals value", geocode)
		}

		if tot == 0.0 {
			return nil, fmt.Errorf("geocode (total) %s: total is 0", geocode)
		}

		result[geocode] = value / tot
	}

	return result, nil
}

// writeCatFile writes a single category file
func writeCatFile(dir string, cat types.Category, values map[types.Geocode]types.Value) error {
	name := filepath.Join(dir, string(cat)+".csv")

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	// XXX the original tile generator doesn't include the category
	// in the header if there are no geos
	headers := []string{"geography_code"}
	if len(values) > 0 {
		headers = append(headers, string(cat))
	}
	if err := w.Write(headers); err != nil {
		return err
	}

	// sort by geography code
	geos := sort.StringSlice{}
	for geo := range values {
		geos = append(geos, string(geo))
	}
	geos.Sort()

	for _, geo := range geos {
		ratio := strconv.FormatFloat(float64(values[types.Geocode(geo)]), 'g', 13, 64)
		if err := w.Write([]string{geo, ratio}); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}
