package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/jszwec/csvutil"
	"github.com/spkg/bom"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const (
	recodes = "recode-lads.csv"

	// geojson property keys
	propCode = "lad17cd"
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
	recodes := flag.String("r", recodes, "input recode csv file")
	flag.Parse()

	dec := json.NewDecoder(os.Stdin)
	var col FeatureCollection
	if err := dec.Decode(&col); err != nil {
		log.Fatal(err)
	}

	newcodes, err := loadrecodes(*recodes)
	if err != nil {
		log.Fatal(err)
	}

	recode(&col, newcodes)

	enc := json.NewEncoder(os.Stdout)
	if err = enc.Encode(col); err != nil {
		log.Fatal(err)
	}
}

func loadrecodes(name string) (newcodes map[string]string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// The LSOA file may contain a UTF-8 BOM
	bomreader := bom.NewReader(f)

	csvreader := csv.NewReader(bomreader)
	dec, err := csvutil.NewDecoder(csvreader)
	if err != nil {
		return nil, err
	}

	type row struct {
		From string `csv:"FromCode"`
		To   string `csv:"ToCode"`
	}

	var rows []row
	if err := dec.Decode(&rows); err != nil {
		return nil, err
	}

	newcodes = map[string]string{}
	for _, row := range rows {
		newcodes[row.From] = row.To
	}

	return newcodes, nil
}

func recode(col *FeatureCollection, newcodes map[string]string) {
	for _, feat := range col.Features {
		code, ok := feat.Properties[propCode].(string)
		if !ok || code == "" {
			continue
		}

		newcode := newcodes[code]
		if newcode == "" {
			continue
		}

		feat.Properties[propCode] = newcode
	}
}
