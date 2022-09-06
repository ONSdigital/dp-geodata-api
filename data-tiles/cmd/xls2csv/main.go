package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tealeg/xlsx"
)

func main() {
	list := flag.Bool("l", false, "list sheets in xlsx")
	sheetname := flag.String("s", "", "name of sheet to extract")
	flag.Parse()

	xls, err := loadXLS(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if *list {
		for _, sheet := range xls.Sheets {
			fmt.Printf("%s\n", sheet.Name)
		}
		os.Exit(1)
	}

	if len(xls.Sheets) == 0 {
		// not sure if this can happen, but just in case
		log.Fatal("spreadsheet has no sheets")
	}

	var sheet *xlsx.Sheet
	if *sheetname != "" {
		var found bool
		sheet, found = xls.Sheet[*sheetname]
		if !found {
			log.Fatal("sheet %s not found", *sheetname)
		}
	} else if len(xls.Sheets) > 1 {
		log.Fatal("must provide sheetname when spreadsheet has multiple sheets")
	} else {
		sheet = xls.Sheets[0]
	}

	if err := extract(sheet, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func loadXLS(r io.Reader) (*xlsx.File, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return xlsx.OpenBinary(buf)
}

func extract(sheet *xlsx.Sheet, w io.Writer) error {
	writer := csv.NewWriter(w)

	for _, row := range sheet.Rows {
		var record []string
		for _, cell := range row.Cells {
			var value string
			if cell == nil { // not sure if this happens
				value = ""
			} else {
				value = cell.String()
			}
			record = append(record, value)
		}
		for len(record) < sheet.MaxCol {
			record = append(record, "")
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
