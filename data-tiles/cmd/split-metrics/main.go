package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/cat"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/types"
)

const (
	srcdir = "data/metrics/downloads"
	dstdir = "data/metrics/final"
)

func main() {
	srcdir := flag.String("s", srcdir, "directory holding .CSV source files")
	dstdir := flag.String("d", dstdir, "directory holding single-category .CSV files")
	pattern := flag.String("p", "*DATA.CSV", "glob pattern to match source .CSV files")
	calcRatios := flag.Bool("R", false, "calculate ratios")
	flag.Parse()

	csvs, err := findcsvs(*srcdir, *pattern)
	if err != nil {
		log.Fatal(err)
	}

	for _, fname := range csvs {
		fmt.Fprintf(os.Stderr, "*")
		f, err := os.Open(fname)
		if err != nil {
			log.Fatal(err)
		}

		csvreader := csv.NewReader(f)
		table, err := csvreader.ReadAll()
		if err != nil {
			log.Fatal(err)
		}
		f.Close()

		if len(table) < 2 {
			log.Printf("%s: no data\n", fname)
			continue
		}

		if err := splitcsv(table, *calcRatios, *dstdir); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintln(os.Stderr)
}

// findcsvs returns a list of files named *DATA.CSV in dir
func findcsvs(dir, pattern string) ([]string, error) {
	var csvs []string

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeType != 0 {
			return nil // must be regular file
		}
		isdata, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return err
		}
		if !isdata {
			return nil
		}
		csvs = append(csvs, path)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return csvs, nil
}

func splitcsv(table [][]string, calcRatios bool, dir string) error {
	if calcRatios {
		return splitWithRatios(table, dir)
	} else {
		return splitWithoutRatios(table, dir)
	}
}

func splitWithoutRatios(table [][]string, dir string) error {
	headings := table[0]

	for col, catcode := range headings {
		if col == 0 {
			continue // skip GeographyCode
		}
		fname := filepath.Join(dir, catcode+".CSV")

		fmt.Fprintf(os.Stderr, ".")
		f, err := os.Create(fname + ".new")
		if err != nil {
			return err
		}

		w := csv.NewWriter(f)
		for _, row := range table {
			if err := w.Write([]string{row[0], row[col]}); err != nil {
				return err
			}
		}

		w.Flush()
		if err := w.Error(); err != nil {
			return err
		}

		f.Close()

		if err := os.Rename(fname+".new", fname); err != nil {
			return err
		}
	}
	return nil
}

func splitWithRatios(table [][]string, dir string) error {
	totals, err := findTotals(table)
	if err != nil {
		return err
	}

	headings := table[0]
	for coln, catcode := range headings {
		if coln == 0 {
			continue
		}
		if cat.IsTotalsCat(types.Category(catcode)) {
			continue
		}
		fname := filepath.Join(dir, catcode+".CSV")
		fmt.Fprintf(os.Stderr, ".")
		f, err := os.Create(fname + ".new")
		if err != nil {
			return err
		}
		w := csv.NewWriter(f)

		if err := genRatioCSV(w, table, coln, totals); err != nil {
			f.Close()
			return err
		}

		w.Flush()
		if err := w.Error(); err != nil {
			f.Close()
			return err
		}
		f.Close()

		if err := os.Rename(fname+".new", fname); err != nil {
			return err
		}
	}
	return nil
}

func findTotals(table [][]string) (map[string]float64, error) {
	for coln, heading := range table[0] {
		if !cat.IsTotalsCat(types.Category(heading)) {
			continue
		}
		totals := map[string]float64{}
		for rown, row := range table {
			if rown == 0 {
				continue
			}
			geocode := row[0]
			val, err := strconv.ParseFloat(row[coln], 64)
			if err != nil {
				return nil, fmt.Errorf("row %d, col %d (%q): %w", rown, coln, row[coln], err)
			}
			if val == 0.0 {
				return nil, fmt.Errorf("row %d, col %d: zero value cannot be used in ratios", rown, coln)
			}
			totals[geocode] = val
		}
		return totals, nil
	}
	return nil, errors.New("no totals column")
}

func genRatioCSV(w *csv.Writer, table [][]string, coln int, totals map[string]float64) error {
	headings := table[0]
	if err := w.Write([]string{headings[0], headings[coln]}); err != nil {
		return err
	}

	for rown, row := range table {
		if rown == 0 {
			continue
		}
		geocode := row[0]
		val, err := strconv.ParseFloat(row[coln], 64)
		if err != nil {
			return fmt.Errorf("row %d, col %d (%q): %w", rown, coln, row[coln], err)
		}
		ratio := strconv.FormatFloat(float64(val/totals[geocode]), 'g', -1, 64)
		if err := w.Write([]string{geocode, ratio}); err != nil {
			return err
		}
	}
	return nil
}
