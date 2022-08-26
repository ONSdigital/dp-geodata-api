package geos

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/types"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

func LoadAll(dir string) (map[types.Geotype]map[types.Geocode]*geom.Bounds, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.geojson"))
	if err != nil {
		return nil, err
	}

	bounds := make(map[types.Geotype]map[types.Geocode]*geom.Bounds)

	for _, fname := range matches {
		log.Printf("Loading %s\n", fname)

		geojson, err := LoadGeojson(fname)
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
			codemap, exists := bounds[types.Geotype(typeprop)]
			if !exists {
				codemap = make(map[types.Geocode]*geom.Bounds)
				bounds[types.Geotype(typeprop)] = codemap
			}
			codemap[types.Geocode(codeprop)] = feat.BBox
		}

	}
	return bounds, nil
}

func LoadGeojson(name string) (*geojson.FeatureCollection, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var col geojson.FeatureCollection
	err = dec.Decode(&col)
	if err != nil {
		return nil, err
	}
	return &col, err
}
