package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/twpayne/go-geom/encoding/geojson"
)

type geo struct {
	En      string    `json:"en"`
	GeoType string    `json:"geoType"`
	GeoCode string    `json:"geoCode"`
	Bbox    []float64 `json:"bbox"`
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(), "usage: %s <geojson-file> ...\n",
			os.Args[0],
		)
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var geos []geo
	for _, name := range flag.Args() {
		col, err := load(name)
		if err != nil {
			log.Fatal(err)
		}

		for _, feat := range col.Features {
			en := feat.Properties["ename"].(string)
			geotype := feat.Properties["geotype"].(string)
			geocode := feat.Properties["geocode"].(string)
			bounds := feat.BBox
			geos = append(
				geos,
				geo{
					En:      en,
					GeoType: geotype,
					GeoCode: geocode,
					Bbox: []float64{
						bounds.Min(0),
						bounds.Min(1),
						bounds.Max(0),
						bounds.Max(1),
					},
				},
			)
		}

	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(geos); err != nil {
		log.Fatal(err)
	}
}

func load(name string) (*geojson.FeatureCollection, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var col geojson.FeatureCollection
	if err := dec.Decode(&col); err != nil {
		return nil, err
	}

	return &col, nil
}
