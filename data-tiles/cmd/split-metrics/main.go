package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const (
	srcdir = "data/metrics/downloads"
	dstdir = "data/metrics/final"
)

func main() {
	srcdir := flag.String("s", srcdir, "directory holding .CSV source files")
	dstdir := flag.String("d", dstdir, "directory holding single-category .CSV files")
	pattern := flag.String("p", "*DATA.CSV", "glob pattern to match source .CSV files/")
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

		if err := splitcsv(table, *dstdir); err != nil {
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

func splitcsv(table [][]string, dir string) error {
	headings := table[0]

	for col, cat := range headings {
		if col == 0 {
			continue // skip GeographyCode
		}

		fname := filepath.Join(dir, cat+".CSV")

		//fmt.Fprintf(os.Stderr, "\t%s\n", cat)
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
