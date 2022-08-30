package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-geom/xy"
)

// The "geojson" output is slightly non-standard.
type GeoOutput struct {
	Meta    Meta                       `json:"meta"`
	GeoJSON *geojson.FeatureCollection `json:"geo_json"`
}

type Meta struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Geotype string `json:"geotype"`
}

func main() {
	codeprop := flag.String("c", "geocode", "geojson property name holding the geocode (eg geocode)")
	enameprop := flag.String("n", "ename", "geojson property name holding the English geography name (eg ename)")
	typeprop := flag.String("t", "geotype", "geojson property name holding the geotype (eg geotype")
	dir := flag.String("d", "data/tiles/geo", "directory to hold output files")
	flag.Parse()

	if *codeprop == "" {
		log.Fatal("must provide name of geography code property (-c)")
	}
	if *enameprop == "" {
		log.Fatal("must provide name of geography name property (-n)")
	}
	if *typeprop == "" {
		log.Fatal("must provide nam of geotype name property (-t)")
	}

	if err := process(*codeprop, *enameprop, *typeprop, *dir); err != nil {
		log.Fatal(err)
	}
}

func process(codeprop, enameprop, typeprop, dir string) error {
	dec := json.NewDecoder(os.Stdin)
	var col geojson.FeatureCollection
	if err := dec.Decode(&col); err != nil {
		return err
	}

	nfeat := len(col.Features)
	stride := nfeat / 100

	for i, feat := range col.Features {
		if i%stride == 0 {
			fmt.Fprintf(os.Stderr, " %d/%d %d%%\r", i, nfeat, (i*100)/nfeat)
		}
		code, ok := feat.Properties[codeprop].(string)
		if !ok || code == "" {
			log.Printf("feature %d: no %s property, skipping", i, codeprop)
			continue
		}

		name, _ := feat.Properties[enameprop].(string)
		geotype, ok := feat.Properties[typeprop].(string)
		if !ok || geotype == "" {
			log.Printf("feature %d: no %s property, skipping", i, typeprop)
			continue
		}

		// calc centroid as a feature, id="centroid"
		p, err := xy.Centroid(feat.Geometry)
		if err != nil {
			log.Printf("feature %d: cannot calculate centroid: %s", i, err)
			continue
		}
		centroid := geom.NewPoint(geom.XY)
		centroid, err = centroid.SetCoords(p)

		// convert bbox to LineString feature id="bbox"
		ls, err := geom.NewLineString(geom.XY).SetCoords(
			[]geom.Coord{
				{feat.BBox.Min(0), feat.BBox.Min(1)},
				{feat.BBox.Max(0), feat.BBox.Max(1)},
			},
		)
		if err != nil {
			log.Printf("feature %d: cannot calc linestring bbox: %s", i, err)
			continue
		}

		out := GeoOutput{
			Meta: Meta{
				Code:    code,
				Name:    name,
				Geotype: geotype,
			},
			GeoJSON: &geojson.FeatureCollection{
				Features: []*geojson.Feature{
					&geojson.Feature{
						ID:       "centroid",
						Geometry: centroid,
					},
					&geojson.Feature{
						ID:       "bbox",
						Geometry: ls,
					},
					&geojson.Feature{
						ID:       "boundary",
						Geometry: feat.Geometry,
					},
				},
			},
		}

		fname := filepath.Join(dir, code+".geojson")
		f, err := os.Create(fname + ".new")
		if err != nil {
			return err
		}

		enc := json.NewEncoder(f)
		enc.SetIndent("", "    ")
		if err = enc.Encode(out); err != nil {
			f.Close()
			return err
		}
		f.Close()

		if err := os.Rename(fname+".new", fname); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, " %d/%d %d%%\n", nfeat, nfeat, 100)
	return nil
}
