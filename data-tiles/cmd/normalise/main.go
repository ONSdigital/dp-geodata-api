package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// geojson.FeatureCollection doesn't have CRS
type FeatureCollection struct {
	CRS      *geojson.CRS       `json:"crs"`
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	BBox     *geom.Bounds       `json:"bbox"`
	Features []*geojson.Feature `json:"features"`
}

func main() {
	codeprop := flag.String("c", "", "property key which holds geocode (eg MSOA11CD)")
	geotype := flag.String("t", "", "geotype to set in each property (eg MSOA)")
	enameprop := flag.String("e", "", "property key which holds English name (eg MSOA11NM)")
	wnameprop := flag.String("w", "", "property key which olds Welsh name (eg MSOA11NMW)")
	flag.Parse()

	dec := json.NewDecoder(os.Stdin)
	var col FeatureCollection
	if err := dec.Decode(&col); err != nil {
		log.Fatal(err)
	}

	for n, feat := range col.Features {
		if *geotype != "" {
			feat.Properties["geotype"] = *geotype
		}
		if *codeprop != "" {
			if err := copyProp(*codeprop, "geocode", feat.Properties); err != nil {
				log.Printf("feature %d: %s", n, err)
			}
		}
		if *enameprop != "" {
			if err := copyProp(*enameprop, "ename", feat.Properties); err != nil {
				log.Printf("feature %d: %s", n, err)
			}
		}
		if *wnameprop != "" {
			if err := copyProp(*wnameprop, "wname", feat.Properties); err != nil {
				log.Printf("feature %d: %s", n, err)
			}

		}
		if feat.BBox == nil {
			feat.BBox = feat.Geometry.Bounds()
		}
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(&col); err != nil {
		log.Fatal(err)
	}
}

func copyProp(src, dst string, props map[string]interface{}) error {
	val, ok := props[src]
	if !ok {
		return fmt.Errorf("property %s not in feature", src)
	}
	props[dst] = val
	return nil
}
