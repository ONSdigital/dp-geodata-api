package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	listflag := flag.Bool("l", false, "list sheets in xlsx")
	sheetname := flag.String("s", "", "name of sheet to extract")
	flag.Parse()

	xls, err := excelize.OpenReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if *listflag {
		err = list(xls)
	} else {
		err = extract(xls, *sheetname)
	}

	if err != nil {
		log.Println(err)
	}
	cerr := xls.Close()
	if cerr != nil {
		log.Println(err)
	}

	if err != nil || cerr != nil {
		os.Exit(1)
	}
}

func list(xls *excelize.File) error {
	aidx := xls.GetActiveSheetIndex()
	asheet := xls.GetSheetName(aidx)

	for _, name := range xls.GetSheetList() {
		var active string
		if name == asheet {
			active = " (active)"
		}
		fmt.Printf("%s%s\n", name, active)
	}
	return nil
}

func extract(xls *excelize.File, sheet string) error {
	if sheet == "" {
		aidx := xls.GetActiveSheetIndex()
		sheet = xls.GetSheetName(aidx)
		if sheet == "" {
			return errors.New("no active sheet, try -s")
		}
	}

	rows, err := xls.GetRows(sheet)
	if err != nil {
		return err
	}

	// count columns
	var cols int
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}

	// extend short rows
	for i, row := range rows {
		for len(row) < cols {
			row = append(row, "")
		}
		rows[i] = row
	}

	// emit csv
	csvw := csv.NewWriter(os.Stdout)
	if err := csvw.WriteAll(rows); err != nil {
		return err
	}
	csvw.Flush()
	return csvw.Error()
}
