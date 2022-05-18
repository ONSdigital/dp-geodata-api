package geodata

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ONSdigital/dp-geodata-api/pkg/timer"
	"github.com/ONSdigital/dp-geodata-api/pkg/where"
	"github.com/ONSdigital/dp-geodata-api/sentinel"
	"github.com/jtrim-ons/ckmeans/pkg/ckmeans"
)

// Implementation notes:
//
// OA-sized datasets are big.
// There are 181408 OAs, which means that SQL queries must return that many rows for each
// geotype-catcode combination.
// An OA ratio query for 7 catcodes, will result in 8 * 181408 rows returned.
//
// The pre-OA implementation sent a single query to postgres, and loaded the results
// into a set of maps and slices, summarising chunk by chunk.
// Unfortunately, with OA-sized queries, sorting in database is a large performance hit,
// and the sheer size of the results consumed way too much memory.
//
// This implementation does two things in an attempt to improve performance and memory
// requirements, but only partially successfully.
//
// 1. Instead of a single large SQL query, we query for each geotype-catcode separately
//    and do not ORDER BY.
//    These queries do not return geotype and catcode as before, but only return geocodes
//    and metric values, reducing the overall size of each result set.
// 2. To avoid large map and slice reallocations, we reuse large maps and slices when
//    possible, especially when loading a geotype-catcode metric set.
//
// Alas, although execution speed was improved, memory usage was not improved enough.
// I still can't get it to run in a 128MB container.
// The sql library (or the underlying pgx library) uses too much memory.
// And the ckmeans library by necessity also uses a measurable amount.
// But our special test query now runs in a 256MB container and comes back in ~3s with a
// local db and ~6.2s against RDS.
// (Previous figures were ~4s and ~10s respectively.)

// CkmeansParams holds data and methods neccessary to parse and process data for ckmeans queries
type CkmeansParams struct {
	year     int
	geotypes []string
	catcodes []string
	divideBy string
	k        int
	db       *sql.DB

	// breaks holds the results to be returned by Ckmeans.
	// The top level map is the category code, and the next level map holds the geotype
	// or <geotype>_min_max.
	// The float64 slice holds the break points or min/max values.
	// For example:
	// 	minmax := map["QS501EW0008"]["OA_min_max"]
	//	(minmax[0] is min, minmax[1] is max)
	// or
	//	breaks := map["QS501EW0008"]["OA"]
	//
	// Certain combinations of errors or missing data call for either nil or an empty map to
	// be returned.
	breaks map[string]map[string][]float64
}

// Ckmeans calculates ckmean breaks and min-max values for the metrics in
// each geotype-catcode combination.
//
// When divideBy is not empty, this is taken to be the denominator of a ratio
// query.
// The metrics of each geotype-catcode combination are divided by the metrics
// of the denominator, matching geocode to geocode.
// ckmeans and min-max are then are calculated over the ratios.
func (app *Geodata) CKmeans(ctx context.Context, year int, cat, geotype []string, k int, divideBy string) (map[string]map[string][]float64, error) {
	catcodes, err := parseCat(cat)
	if err != nil {
		return nil, err
	}

	geotypes, err := parseValidateGeotype(geotype)
	if err != nil {
		return nil, err
	}

	params := &CkmeansParams{
		year:     year,
		catcodes: catcodes,
		geotypes: geotypes,
		divideBy: divideBy,
		k:        k,
		db:       app.db.DB(),
		breaks:   map[string]map[string][]float64{},
	}

	if divideBy == "" {
		err = params.nonratio(ctx)
	} else {
		err = params.ratio(ctx)
	}
	if err != nil {
		params.breaks = nil
	}
	return params.breaks, err
}

// nonratio calculates ckmeans over geotype-category metrics directly.
func (params *CkmeansParams) nonratio(ctx context.Context) error {
	values := []float64{}
	metrics := map[string]float64{}
	for _, geotype := range params.geotypes {
		for _, catcode := range params.catcodes {
			if err := params.loadMetrics(ctx, geotype, catcode, metrics); err != nil {
				return err
			}
			if len(metrics) == 0 {
				params.breaks = map[string]map[string][]float64{} // special case
				return nil
			}

			values = values[:0] // reuse existing slice
			for _, value := range metrics {
				values = append(values, value)
			}

			if err := params.collectStats(values, geotype, catcode); err != nil {
				return err
			}
		}
	}
	return nil
}

// ratio calculates ckmeans over the ratio of geotype-category metrics to the
// metrics of geotype-divideBy.
func (params *CkmeansParams) ratio(ctx context.Context) error {
	values := []float64{}
	denominator := map[string]float64{}
	numerator := map[string]float64{}
	for _, geotype := range params.geotypes {
		if err := params.loadMetrics(ctx, geotype, params.divideBy, denominator); err != nil {
			return err
		}
		if len(denominator) == 0 {
			params.breaks = map[string]map[string][]float64{} // special case
			return nil
		}

		for _, catcode := range params.catcodes {
			if err := params.loadMetrics(ctx, geotype, catcode, numerator); err != nil {
				return err
			}
			if len(numerator) == 0 || len(numerator) != len(denominator) {
				return fmt.Errorf("%w: %s %s", sentinel.ErrPartialContent, geotype, catcode)
			}

			values = values[:0] // reuse existing slice
			for geocode, d := range denominator {
				if d == 0 {
					return fmt.Errorf("%w: %s %s %s == 0", sentinel.ErrInvalidParams, geotype, catcode, geocode)
				}
				n, ok := numerator[geocode]
				if !ok {
					return fmt.Errorf("%w: %s %s %s", sentinel.ErrPartialContent, geotype, catcode, geocode)
				}
				values = append(values, n/d)
			}

			if err := params.collectStats(values, geotype, catcode); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectStats calculates statistics on metrics (ckmeans, min, max) and saves the results
// against geotype and catcode.
func (params *CkmeansParams) collectStats(metrics []float64, geotype, catcode string) error {
	catBreaks, err := getBreaks(metrics, params.k)
	if err != nil {
		return err
	}

	cc, ok := params.breaks[catcode]
	if !ok {
		cc = map[string][]float64{}
	}
	cc[geotype] = catBreaks
	cc[geotype+"_min_max"] = getMinMax(metrics)
	params.breaks[catcode] = cc
	return nil
}

var ckquery = `
SELECT
	geo.code AS geography_code,
	geo_metric.metric AS value
FROM
	geo,
	geo_type,
	geo_metric,
	data_ver,
	nomis_category
WHERE geo.valid
AND geo_type.id = geo.type_id
AND geo_type.name = $1
AND geo_metric.geo_id = geo.id
AND data_ver.id = geo_metric.data_ver_id
AND data_ver.census_year = $2
AND data_ver.ver_string = '2.2'
AND nomis_category.id = geo_metric.category_id
AND nomis_category.year = data_ver.census_year
AND nomis_category.long_nomis_code = $3
`

// loadMetrics retrieves metrics for geotype and catcode and places them in result.
// Result keys are geocodes and the metric values are the map values.
func (params *CkmeansParams) loadMetrics(ctx context.Context, geotype, catcode string, result map[string]float64) error {
	var err error
	rows, err := params.db.QueryContext(ctx, ckquery, geotype, params.year, catcode)
	if err != nil {
		return err
	}
	defer rows.Close()

	var k string
	for k = range result {
		delete(result, k)
	}
	var geocode string
	var value float64
	for rows.Next() {
		if err = rows.Scan(&geocode, &value); err != nil {
			return err
		}
		result[geocode] = value
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// parseCat parses single values and combines with split comma-seperated cat values and returns as array.
// Will return error if any cat range values (cat1..cat2) are found (for error handling we need to know
// explicitly beforehand which cats to expect in our results)
//
func parseCat(cat []string) ([]string, error) {
	catset, err := where.ParseMultiArgs(cat)
	if err != nil {
		return nil, err
	}
	if catset.Ranges != nil {
		return nil, fmt.Errorf("%w: ckmeans endpoint does not accept range values for cats", sentinel.ErrInvalidParams)
	}
	return catset.Singles, nil
}

// parseValidateGeotype parses single values and combines with split comma-seperated geotype values
// and returns as array. Any badly-cased geotype values will be corrected, and any unrecognised geotype
// value will cause an error to be returned.
//
func parseValidateGeotype(geotype []string) ([]string, error) {
	geoset, err := where.ParseMultiArgs(geotype)
	if err != nil {
		return nil, err
	}
	geoset, err = MapGeotypes(geoset)
	if err != nil {
		return nil, err
	}
	return geoset.Singles, nil
}

// getBreaks gets k ckmeans clusters from metrics and returns the upper breakpoints for each cluster.
//
func getBreaks(metrics []float64, k int) ([]float64, error) {
	clusters, err := ckmeans.Ckmeans(metrics, k)
	if err != nil {
		return nil, err
	}

	var breaks []float64
	for _, cluster := range clusters {
		bp := cluster[len(cluster)-1]
		breaks = append(breaks, bp)
	}
	return breaks, nil
}

func getMinMax(values []float64) []float64 {
	max := values[0]
	min := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	return []float64{min, max}
}

// !!!! DEPRECATED CKMEANSRATIO TO BE REMOVED WHEN FRONT END REMOVES DEPENDENCY ON IT !!!!
//
func (app *Geodata) CKmeansRatio(ctx context.Context, year int, cat1 string, cat2 string, geotype string, k int) ([]float64, error) {
	sql := `
SELECT
    geo_metric.metric
	, nomis_category.long_nomis_code
	, geo.id
FROM
    geo,
    geo_type,
    nomis_category,
    geo_metric,
    data_ver
-- the geo_type we are interested in
WHERE geo_type.name = $1 
-- all geocodes in this type
AND geo.type_id = geo_type.id
AND geo.valid
-- the category we are interested in
AND nomis_category.long_nomis_code IN ($2, $3)
AND nomis_category.year = $4
-- metrics for these geocodes and category
AND geo_metric.geo_id = geo.id
AND geo_metric.category_id = nomis_category.id
-- only pick metrics for census year / version2.2
AND data_ver.id = geo_metric.data_ver_id
AND data_ver.census_year = nomis_category.year
AND data_ver.ver_string = '2.2'
`

	t := timer.New("query")
	t.Start()
	rows, err := app.db.DB().QueryContext(
		ctx,
		sql,
		geotype,
		cat1,
		cat2,
		year,
	)

	if err != nil {
		return nil, err
	}
	t.Stop()
	t.Log(ctx)
	defer rows.Close()

	tnext := timer.New("next")
	tscan := timer.New("scan")

	var nmetricsCat1 int
	var nmetricsCat2 int
	metricsCat1 := make(map[int]float64)
	metricsCat2 := make(map[int]float64)
	for {
		tnext.Start()
		ok := rows.Next()
		tnext.Stop()
		if !ok {
			break
		}

		var (
			metric float64
			cat    string
			geoID  int
		)
		tscan.Start()
		err := rows.Scan(&metric, &cat, &geoID)
		tscan.Stop()
		if err != nil {
			return nil, err
		}
		if cat == cat1 {
			nmetricsCat1++
			metricsCat1[geoID] = metric
		}
		if cat == cat2 {
			nmetricsCat2++
			metricsCat2[geoID] = metric
		}
	}
	tnext.Log(ctx)
	tscan.Log(ctx)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if nmetricsCat1 == 0 && nmetricsCat2 == 0 {
		return []float64{}, nil
	}
	if nmetricsCat1 != nmetricsCat2 {
		return nil, sentinel.ErrPartialContent
	}

	var metrics []float64
	for geoID, metricCat1 := range metricsCat1 {
		metricCat2, prs := metricsCat2[geoID]
		if !prs {
			return nil, sentinel.ErrPartialContent
		}
		metrics = append(metrics, metricCat1/metricCat2)
	}

	return getBreaks(metrics, k)
}
