package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ONSdigital/dp-geodata-api/model"
	"github.com/ONSdigital/dp-geodata-api/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const MAX = 8

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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dsn := database.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Print(err)
	}

	fmt.Println(dsn)

	t0 := time.Now()
	parsePostcodeCSV(db, "PCD_OA_LSOA_MSOA_LAD_MAY20_UK_LU.csv")
	fmt.Printf("%d min(s)\n", int(time.Since(t0).Minutes()))
}

func parsePostcodeCSV(db *gorm.DB, file string) {

	records := readCsvFile(file)

	field := make(map[string]int)
	for i, k := range records[0] {
		field[k] = i
	}

	var j int32

	sem := make(chan int, MAX)

	wg := new(sync.WaitGroup)
	for i, line := range records {
		if i == 0 {
			continue
		}

		pcds := line[field["pcds"]]
		msoa11cd := line[field["msoa11cd"]]
		// Scotland isn't of interest nor Northern Ireland nor Channel Islands nor Isle of Man
		if !strings.HasPrefix(msoa11cd, "S") && !strings.HasPrefix(msoa11cd, "N") && !strings.HasPrefix(msoa11cd, "L") && !strings.HasPrefix(msoa11cd, "M") && msoa11cd != "" {

			sem <- 1
			wg.Add(1)
			go func(pcds, msoa11cd string) {
				defer func() {
					wg.Done()
					<-sem
				}()
				var geos model.Geo
				db.Where("type_id = 5 and code=?", msoa11cd).Find(&geos) // limit by MSOA
				if geos.ID == 0 {
					log.Fatalf("not found: %s", msoa11cd)
				}

				var pc model.PostCode
				pc.GeoID = geos.ID
				pc.Pcds = pcds
				db.Save(&pc)

				atomic.AddInt32(&j, 1)

				if j%100000 == 0 {
					fmt.Printf("~%.1f%% ... ", (float64(j)/2300000)*100)
				}
			}(pcds, msoa11cd)
		}

	}

	wg.Wait()
	fmt.Printf("%d rows\n", j)
}
