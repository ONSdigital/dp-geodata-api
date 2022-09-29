package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/spkg/bom"
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
	names := flag.String("n", "", "input MSOA names file")
	chdr := flag.String("c", "msoa11cd", "geocode CSV header")
	ehdr := flag.String("e", "msoa11hclnm", "English name CSV header")
	whdr := flag.String("w", "msoa11hclnmw", "Welsh name CSV header")
	cprop := flag.String("C", "MSOA11CD", "geocode geojson property")
	eprop := flag.String("E", "MSOA11NM", "English name geojson property")
	wprop := flag.String("W", "MSOA11NMW", "Welsh name geojson property")
	flag.Parse()

	if *names == "" {
		log.Fatal("must provide MSOA names file (-n)")
	}

	if *chdr == "" || *ehdr == "" || *whdr == "" || *cprop == "" || *eprop == "" || *wprop == "" {
		log.Fatal("must provide all header and property names")
	}

	enames, wnames, err := loadnames(*names, *chdr, *ehdr, *whdr)
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(os.Stdin)
	var col FeatureCollection
	if err := dec.Decode(&col); err != nil {
		log.Fatal(err)
	}

	rename(&col, enames, wnames, *cprop, *eprop, *wprop)

	enc := json.NewEncoder(os.Stdout)
	if err = enc.Encode(col); err != nil {
		log.Fatal(err)
	}
}

func loadnames(name, codehdr, ehdr, whdr string) (enames, wnames map[string]string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// The MSOA file may contain a UTF-8 BOM
	bomreader := bom.NewReader(f)

	csvreader := csv.NewReader(bomreader)

	rows, err := csvreader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(rows) <= 1 {
		return nil, nil, errors.New("not enough data in MSOA names file")
	}

	ccol := findHeader(codehdr, rows[0])
	ecol := findHeader(ehdr, rows[0])
	wcol := findHeader(whdr, rows[0])
	if ccol == -1 || ecol == -1 || wcol == -1 {
		return nil, nil, errors.New("could not find headers in MSOA names file")
	}

	enames = map[string]string{}
	wnames = map[string]string{}
	for _, row := range rows[1:] {
		geocode := row[ccol]
		ename := row[ecol]
		if ename != "" {
			enames[geocode] = ename
		}
		wname := row[wcol]
		if wname != "" {
			wnames[geocode] = wname
		}
	}

	return enames, wnames, nil
}

func findHeader(want string, row []string) int {
	for i, header := range row {
		if header == want {
			return i
		}
	}
	return -1
}

func rename(col *FeatureCollection, enames, wnames map[string]string, cprop, eprop, wprop string) {
	for _, feat := range col.Features {
		code, ok := feat.Properties[cprop].(string)
		if !ok || code == "" {
			continue
		}

		ename := enames[code]
		wname := wnames[code]
		if wname == "" {
			wname = ename
		}

		if ename != "" {
			feat.Properties[eprop] = ename
		}

		if wname != "" {
			feat.Properties[wprop] = wname
		}
	}
}
