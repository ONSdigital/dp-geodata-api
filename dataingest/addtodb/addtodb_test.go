//go:build comptest
// +build comptest

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/ONSdigital/dp-geodata-api/comptests"
	"github.com/ONSdigital/dp-geodata-api/model"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dsn = comptests.DefaultDSN

var db *gorm.DB

func init() {
	comptests.SetupDockerDB(dsn)
	model.SetupUpdateDB(dsn)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

}

type qLogger struct {
}

func (l *qLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	// uncomment me for logs
	//	fmt.Printf("SQL:\n%s\nARGS:%v\n", data["sql"], data["args"])
}

func TestGetFiles(t *testing.T) {
	di := New("2011", dsn)

	di.getFiles("testdata/")

	for _, fn := range di.files.data {
		if fn != "testdata/QS104EWDATA01_A.CSV" && fn != "testdata/QS104EWDATA04.CSV" {
			t.Fail()
		}
	}

	if di.files.meta[0] != "testdata/QS104EWMETA0.CSV" || di.files.desc[0] != "testdata/QS104EWDESC0.CSV" {
		t.Fail()
	}
}

func TestAddClassificationData(t *testing.T) {

	func() {
		var nd model.NomisDesc
		tx := db.Begin()
		defer tx.Rollback()

		di := New("2011", dsn)
		di.gdb = tx
		di.files.meta = []string{"testdata/QS104EWMETA0.CSV"}

		if foo := tx.First(&nd); !errors.Is(foo.Error, gorm.ErrRecordNotFound) {
			t.Errorf("Data wrongly present")
		}

		di.addClassificationData()

		tx.First(&nd)

		if nd.Name != "Sex" || nd.PopStat != "All usual residents" || nd.ShortNomisCode != "QS104EW" {
			t.Errorf(fmt.Sprintf("wrongly got : %#v", nd))
		}
	}()

}

func TestAddCategoryData(t *testing.T) {

	func() {
		tx := db.Begin()
		defer tx.Rollback()

		di := New("2011", dsn)
		di.gdb = tx
		di.files.meta = []string{"testdata/QS104EWMETA0.CSV"}
		di.addClassificationData()
		di.files.desc = []string{"testdata/QS104EWDESC0.CSV"}
		longToCatid := di.addCategoryData()

		if longToCatid["QS104EW0001"] == 0 || longToCatid["QS104EW0002"] == 0 {
			t.Error("data not there")
		}

		fmt.Printf("%#v\n", longToCatid)
	}()
}

func TestAddGeoGeoMetricData(t *testing.T) {
	ctx := context.Background()

	func() {
		di := New("2011", dsn)

		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Print(err)
		}

		config.ConnConfig.Logger = &qLogger{}
		pool, err := pgxpool.ConnectConfig(context.Background(), config)
		if err != nil {
			log.Fatal(err)
		}

		// can't get rollback to work with pool
		// manual rollback below

		/*
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Error(err)
			}
			defer tx.Rollback(ctx)

		*/

		di.pool = pool

		pool.Exec(ctx, "INSERT INTO geo_type VALUES(4,'LAD')")
		pool.Exec(ctx, "INSERT INTO geo_type VALUES(7,'OA')")
		pool.Exec(ctx, "INSERT INTO NOMIS_DESC (id,name,pop_stat,short_nomis_code,year,nomis_topic_id) VALUES (66,'Sex','All usual residents','QS104EW',2011,1)")
		pool.Exec(ctx, "INSERT INTO NOMIS_CATEGORY (id,nomis_desc_id,category_name,measurement_unit,stat_unit,long_nomis_code,year) VALUES (3,66,'All categories: Sex','Count','Person','QS104EW0001',2011)")
		pool.Exec(ctx, "INSERT INTO NOMIS_CATEGORY (id,nomis_desc_id,category_name,measurement_unit,stat_unit,long_nomis_code,year) VALUES (4,66,'All categories: Sex','Count','Person','QS104EW0002',2011)")

		di.files.data = []string{"testdata/QS104EWDATA04.CSV", "testdata/QS104EWDATA01_A.CSV"}
		di.addGeoGeoMetricData(map[string]int32{"QS104EW0001": 3, "QS104EW0002": 4})

		var returned float64
		var expected float64

		// check LAD result for QS104EW0001 matches
		if err := pool.QueryRow(
			ctx,
			`
SELECT
	geo_metric.metric
FROM
	geo_metric,
	geo
WHERE geo_metric.geo_id = geo.id
AND geo.type_id = 4
AND geo_metric.category_id = 3
`,
		).Scan(&returned); err != nil {
			t.Error(err)
		}

		expected = 92028
		if returned != expected {
			t.Logf(fmt.Sprintf("Expected %v value for test LAD, got %v", expected, returned))
			t.Fail()
		}

		// check OA result for QS104EW0002 matches
		if err := pool.QueryRow(
			ctx,
			`
SELECT
	geo_metric.metric 
FROM
	geo_metric,
	geo
WHERE geo_metric.geo_id = geo.id
AND geo.type_id = 7
AND geo_metric.category_id = 4
`,
		).Scan(&returned); err != nil {
			t.Error(err)
		}
		expected = 54751
		if returned != expected {
			t.Logf(fmt.Sprintf("Expected %v value for test OA, got %v", expected, returned))
			t.Fail()
		}

		// manual rollback :-/
		pool.Exec(ctx, "DELETE FROM geo_metric")
		pool.Exec(ctx, "DELETE FROM geo")
		pool.Exec(ctx, "DELETE FROM geo_type")
		pool.Exec(ctx, "DELETE FROM NOMIS_CATEGORY")
		pool.Exec(ctx, "DELETE FROM NOMIS_DESC")

	}()

}

func TestGeotypeIDFromGSSCode(t *testing.T) {
	di := New("2011", dsn)
	testGSSCodes := map[string]int{
		// EW test
		"K04000001": 1,
		// Country tests
		"E92000001": 2,
		"W92000004": 2,
		// region tests
		"E12000004": 3,
		// LAD tests,
		"E06000027": 4,
		"W06000023": 4,
		"E07000009": 4,
		"E08000007": 4,
		"E09000003": 4,
		// MSOA tests
		"E02000030": 5,
		"W02000405": 5,
		// LSOA tests
		"E01000019": 6,
		"W01001952": 6,
		// OA tests
		"E00080972": 7,
		"W00000025": 7,
	}
	for gssCode, expected := range testGSSCodes {
		returned, err := di.geotypeIDFromGSSCode(gssCode)
		if err != nil {
			t.Error(err)
		}
		if returned != expected {
			t.Logf(fmt.Sprintf("Expected geotype ID %d for GSS code %s, got %d", expected, gssCode, returned))
			t.Fail()
		}
	}
}
