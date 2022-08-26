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
	names = "data/downloads/geo/MSOA-Names-1.16.csv"

	// geojson property keys
	propCode  = "MSOA11CD"
	propEname = "MSOA11NM"
	propWname = "MSOA11NMW"
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
	names := flag.String("n", names, "input MSOA names file")
	flag.Parse()

	dec := json.NewDecoder(os.Stdin)
	var col FeatureCollection
	if err := dec.Decode(&col); err != nil {
		log.Fatal(err)
	}

	enames, wnames, err := loadnames(*names)
	if err != nil {
		log.Fatal(err)
	}

	rename(&col, enames, wnames)

	enc := json.NewEncoder(os.Stdout)
	if err = enc.Encode(col); err != nil {
		log.Fatal(err)
	}
}

func loadnames(name string) (enames, wnames map[string]string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// The MSOA file contains a UTF-8 BOM
	bomreader := bom.NewReader(f)

	csvreader := csv.NewReader(bomreader)
	dec, err := csvutil.NewDecoder(csvreader)
	if err != nil {
		return nil, nil, err
	}

	type hcname struct {
		Code          string `csv:"msoa11cd"`
		EnglishName   string `csv:"msoa11nm"`
		WelshName     string `csv:"msoa11nmw"`
		HCEnglishName string `csv:"msoa11hclnm"`
		HCWelshName   string `csv:"msoa11hclnmw"`
		Laname        string `csv:"Laname"`
	}

	var hcnames []hcname
	if err := dec.Decode(&hcnames); err != nil {
		return nil, nil, err
	}

	enames = map[string]string{}
	wnames = map[string]string{}
	for _, row := range hcnames {
		if row.HCEnglishName != "" {
			enames[row.Code] = row.HCEnglishName
		}
		if row.HCWelshName != "" {
			wnames[row.Code] = row.HCWelshName
		}
	}

	return enames, wnames, nil
}

func rename(col *FeatureCollection, enames, wnames map[string]string) {
	for _, feat := range col.Features {
		code, ok := feat.Properties[propCode].(string)
		if !ok || code == "" {
			continue
		}

		ename := enames[code]
		wname := wnames[code]
		if wname == "" {
			wname = ename
		}

		if ename != "" {
			feat.Properties[propEname] = ename
		}

		if wname != "" {
			feat.Properties[propWname] = wname
		}
	}
}
