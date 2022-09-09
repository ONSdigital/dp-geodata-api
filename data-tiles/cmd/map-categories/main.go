package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/content"
)

func main() {
	cfile := flag.String("c", "content.json", "path to content.json")
	flag.Parse()

	c, err := content.LoadName(*cfile)
	if err != nil {
		log.Fatal(err)
	}

	records, err := loadSpreadsheet(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	cmap := buildMap(c)
	headings, keepcols := convertHeadings(records[0], cmap)

	cwriter := csv.NewWriter(os.Stdout)
	if err := cwriter.Write(headings); err != nil {
		log.Fatal(err)
	}

	for _, record := range records[1:] {
		var newrecord []string
		for _, col := range keepcols {
			newrecord = append(newrecord, record[col])
		}
		if err := cwriter.Write(newrecord); err != nil {
			log.Fatal(err)
		}
	}

	cwriter.Flush()
	if err := cwriter.Error(); err != nil {
		log.Fatal(err)
	}
}

// loadSpreadsheet loads the metrics CSV.
func loadSpreadsheet(f io.Reader) ([][]string, error) {
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("empty spreadsheet")
	}
	return records, nil
}

// buildMap creates the map from category name to category code.
func buildMap(c *content.Content) map[string]string {
	catmap := map[string]string{}

	for _, group := range c.TopicGroups {
		for _, topic := range group.Topics {
			for _, variable := range topic.Variables {
				for _, classification := range variable.Classifications {
					for _, category := range classification.Categories {
						catmap[category.Name] = category.Code
					}
				}
			}
		}
	}
	return catmap
}

// convertHeadings renames headings that have mappings, and records which
// columns we want to include in the output spreadsheet.
func convertHeadings(headings []string, cmap map[string]string) ([]string, []int) {
	var newheadings []string
	var keepcols []int

	for i, heading := range headings {
		newheading := lookupHeading(heading, cmap)
		if newheading == "" {
			continue
		}
		newheadings = append(newheadings, newheading)
		keepcols = append(keepcols, i)
	}

	return newheadings, keepcols
}

// lookupHeading maps a single heading, or returns "" if the heading should not
// or could not be mapped.
func lookupHeading(header string, cmap map[string]string) string {
	if header == "geog_code" {
		return "GeographyCode"
	}
	if header == "geog_label" {
		return ""
	}

	// split year and description
	year, desc, found := strings.Cut(header, " ")
	if !found {
		return ""
	}
	if year != "2021" {
		return ""
	}

	// lookup description in map
	newheader, found := cmap[desc]
	if !found {
		return ""
	}

	return year + "-" + newheader
}
