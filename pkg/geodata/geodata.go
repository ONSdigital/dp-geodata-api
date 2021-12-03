package geodata

import (
	"context"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-find-insights-poc-api/pkg/database"
	"github.com/ONSdigital/dp-find-insights-poc-api/pkg/table"
	"github.com/ONSdigital/dp-find-insights-poc-api/pkg/timer"
	"github.com/ONSdigital/dp-find-insights-poc-api/pkg/where"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
)

type Geodata struct {
	db         *database.Database
	maxMetrics int
}

func New(db *database.Database, maxMetrics int) (*Geodata, error) {
	return &Geodata{
		db:         db,
		maxMetrics: maxMetrics,
	}, nil
}

func (app *Geodata) Query(ctx context.Context, dataset, bbox, geotype string, rows, cols []string) (string, error) {
	if len(bbox) > 0 {
		return app.bboxQuery(ctx, bbox, geotype, cols)
	}
	return app.rowQuery(ctx, rows, cols)
}

// rowQuery returns the csv table for the given geometry and category codes.
//
func (app *Geodata) rowQuery(ctx context.Context, geos, cats []string) (string, error) {

	if len(geos) == 0 && len(cats) == 0 {
		return "", ErrMissingParams
	}

	// Construct SQL
	//
	template := `
SELECT
    geo.code AS geography_code,
    nomis_category.long_nomis_code AS category_code,
    geo_metric.metric AS value
FROM
    geo_metric,
    geo,
    nomis_category,
    data_ver
WHERE data_ver.census_year = 2011
AND data_ver.ver_string = '2.2'
%s
AND geo_metric.data_ver_id = data_ver.id
AND geo_metric.geo_id = geo.id
%s
AND nomis_category.year = %d
AND geo_metric.category_id = nomis_category.id
`

	geoWhere, err := additionalCondition("geo.code", geos)
	if err != nil {
		return "", err
	}

	catWhere, err := additionalCondition("nomis_category.long_nomis_code", cats)
	if err != nil {
		return "", err
	}

	sql := fmt.Sprintf(
		template,
		geoWhere,
		catWhere,
		2011,
	)
	fmt.Printf("sql: %s\n", sql)

	return app.collectCells(ctx, sql)
}

// bboxQuery returns the csv table for LSOAs intersecting with the given bbox
//
func (app *Geodata) bboxQuery(ctx context.Context, bbox, geotype string, cats []string) (string, error) {
	var p1lat, p1lon, p2lat, p2lon float64
	fields, err := fmt.Sscanf(bbox, "%f,%f,%f,%f", &p1lat, &p1lon, &p2lat, &p2lon)
	if err != nil {
		return "", fmt.Errorf("scanning bbox %q: %w", bbox, err)
	}
	if fields != 4 {
		return "", fmt.Errorf("bbox missing a number: %w", ErrMissingParams)
	}
	if geotype == "" {
		return "", fmt.Errorf("geotype required: %w", ErrMissingParams)
	}

	// Construct SQL
	//
	template := `
SELECT
	geo.code AS geography_code,
	nomis_category.long_nomis_code AS category_code,
	geo_metric.metric AS value
FROM
	geo,
	geo_type,
	geo_metric,
	data_ver,
	nomis_category
WHERE geo.wkb_geometry && ST_GeomFromText(
		'MULTIPOINT(%f %f, %f %f)',
		4326
	)
AND geo.valid
AND geo.type_id = geo_type.id
AND geo_type.name = %s
AND geo_metric.geo_id = geo.id
AND data_ver.id = geo_metric.data_ver_id
AND data_ver.census_year = %d
AND data_ver.ver_string = '2.2'
AND nomis_category.id = geo_metric.category_id
AND nomis_category.year = %d
%s
`

	catWhere, err := additionalCondition("nomis_category.long_nomis_code", cats)
	if err != nil {
		return "", err
	}

	sql := fmt.Sprintf(
		template,
		p1lon,
		p1lat,
		p2lon,
		p2lat,
		pq.QuoteLiteral(geotype),
		2011,
		2011,
		catWhere,
	)

	fmt.Printf("sql: %s\n", sql)

	return app.collectCells(ctx, sql)
}

// collectCells runs the query in sql and returns the results as a csv.
// sql must be a query against the geo_metric table selecting exactly
// code, category and metric.
//
func (app *Geodata) collectCells(ctx context.Context, sql string) (string, error) {
	// Allocate output table
	//
	tbl, err := table.New("geography_code")
	if err != nil {
		return "", err
	}

	// Set up output buffer
	//
	var body strings.Builder
	body.Grow(1000000)

	// Query the db.
	//
	t := timer.New("query")
	t.Start()
	rows, err := app.db.DB().QueryContext(ctx, sql)
	if err != nil {
		return "", err
	}
	t.Stop()
	t.Print()
	defer rows.Close()

	tnext := timer.New("next")
	tscan := timer.New("scan")
	var nmetrics int
	for {
		tnext.Start()
		ok := rows.Next()
		tnext.Stop()
		if !ok {
			break
		}

		if app.maxMetrics > 0 {
			nmetrics++
			if nmetrics > app.maxMetrics {
				return "", ErrTooManyMetrics
			}
		}

		var geo string
		var cat string
		var value float64

		tscan.Start()
		err := rows.Scan(&geo, &cat, &value)
		tscan.Stop()
		if err != nil {
			return "", err
		}

		tbl.SetCell(geo, cat, value)
	}
	tnext.Print()
	tscan.Print()

	if err := rows.Err(); err != nil {
		return "", err
	}

	tgen := timer.New("generate")
	tgen.Start()
	err = tbl.Generate(&body)
	tgen.Stop()
	tgen.Print()
	if err != nil {
		return "", err
	}

	return body.String(), nil
}

// additionalCondition wraps the output of WherePart inside "AND (...)".
// We "know" this additionalCondition will not be the first additionalCondition in the query.
func additionalCondition(col string, args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	body, err := where.WherePart(col, args)
	if err != nil {
		return "", err
	}

	template := `
AND (
%s
)`
	return fmt.Sprintf(template, body), nil
}
