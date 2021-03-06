package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/ONSdigital/dp-geodata-api/model"
	"github.com/ONSdigital/dp-geodata-api/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func readCsvFile(filePath string) (records [][]string) {
	func() {
		f, err := os.Open(filePath)
		if err != nil {
			log.Fatal("Unable to read input file "+filePath,
				err)
		}
		defer f.Close()

		csvReader := csv.NewReader(f)
		records, err = csvReader.ReadAll()

		if err != nil {
			log.Fatal("Unable to parse file as CSV for "+
				filePath, err)
		}
	}()

	return records
}

func main() {
	dsn := database.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Print(err)
	}

	fmt.Println(dsn)

	parseMsoaCSV(db, "MSOA-Names-1.16.csv")
}

func parseMsoaCSV(db *gorm.DB, file string) {

	records := readCsvFile(file)
	m := make(map[string]string)

	for _, r := range records {
		if r[1] != "" {
			m[r[0]] = r[3] // RHS field 4 msoa11hclnm
		}
	}

	var geos []model.Geo

	db.Where("type_id = 5").Find(&geos) // limit by MSOA

	for i := range geos {
		g := geos[i]

		if m[g.Code] != "" {
			fmt.Print(g.Code)
			fmt.Print(" ")
			fmt.Println(m[g.Code])
			g.Name = m[g.Code]
			db.Save(&g)
		}

	}
}
