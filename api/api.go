// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.2 DO NOT EDIT.
package api

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/go-chi/chi/v5"
)

// Categories defines model for Categories.
type Categories []Triplet

// Error defines model for Error.
type Error struct {
	// error message
	Error string `json:"error"`
}

// Metadata defines model for Metadata.
type Metadata struct {
	Code   *string `json:"code,omitempty"`
	Name   *string `json:"name,omitempty"`
	Slug   *string `json:"slug,omitempty"`
	Tables *Tables `json:"tables,omitempty"`
}

// MetadataResponse defines model for MetadataResponse.
type MetadataResponse []Metadata

// Table defines model for Table.
type Table struct {
	Categories *Categories `json:"categories,omitempty"`
	Code       *string     `json:"code,omitempty"`
	Name       *string     `json:"name,omitempty"`
	Slug       *string     `json:"slug,omitempty"`
	Total      *Triplet    `json:"total,omitempty"`
}

// Tables defines model for Tables.
type Tables []Table

// Triplet defines model for Triplet.
type Triplet struct {
	Code *string `json:"code,omitempty"`
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

// GetCkmeansYearParams defines parameters for GetCkmeansYear.
type GetCkmeansYearParams struct {
	// The census data category to calculate data breaks for.
	// (NB - use metadata endpoint to see list of currently available census data).
	// Can be:
	//   - single values (e.g. QS202EW0002)
	//   - comma-separated array of values (e.g QS202EW0003,QS202EW0003,QS202EW0004)
	//
	// Multiple cats parameters can be supplied, e.g. cats=QS202EW0002&cats=QS202EW0003
	// NB - use of ranges (e.g. QS202EW0003...QS202EW0004) is NOT supported for the ckmeans endpoint.
	Cat *[]string `json:"cat,omitempty"`

	// The type of geography to calculate data breaks for.
	// At the moment these options are supported:
	//   - LAD
	//   - LSOA
	//
	// Can be:
	//   - single values (e.g. LAD)
	//   - comma-separated array of values (e.g LAD,LSOA)
	//
	// Multiple geotype parameters can be supplied, e.g. geotype=LAD&geotype=LSOA
	Geotype *[]string `json:"geotype,omitempty"`

	// The number of data breaks to estimate.
	K *int `json:"k,omitempty"`

	// (OPTIONAL) - census data category to use as denominator (cat/divide_by) to ratios for calculating data
	// breaks, instead of raw data (NB if multiple cat are supplied, each cat will be divided by divide_by). Only
	// single values for divide_by are supported.
	DivideBy *string `json:"divide_by,omitempty"`
}

// GetCkmeansratioYearParams defines parameters for GetCkmeansratioYear.
type GetCkmeansratioYearParams struct {
	// The census data category to use as numerator (cat1/cat2) when producing the ratio to calculate data breaks for
	// (NB - use metadata endpoint to see list of currently available census data).
	Cat1 *string `json:"cat1,omitempty"`

	// The census data category to use as denominator (cat1/cat2) when producing the ratio to calculate data breaks for
	// (NB - use metadata endpoint to see list of currently available census data).
	Cat2 *string `json:"cat2,omitempty"`

	// The type of geography to calculate data breaks for.
	Geotype *string `json:"geotype,omitempty"`

	// The number of data breaks to estimate.
	K *int `json:"k,omitempty"`
}

// GetGeoParams defines parameters for GetGeo.
type GetGeoParams struct {
	// Geography code, eg E09000004
	Geocode *string `json:"geocode,omitempty"`

	// Geography name, eg Bexley
	Geoname *string `json:"geoname,omitempty"`
}

// GetMetadataYearParams defines parameters for GetMetadataYear.
type GetMetadataYearParams struct {
	// Use filtertotals=true if you want to have 'totals' categories separated from other categories in the response (see Examples).
	Filtertotals *bool `json:"filtertotals,omitempty"`
}

// GetQueryYearParams defines parameters for GetQueryYear.
type GetQueryYearParams struct {
	// [ONS codes](https://en.wikipedia.org/wiki/ONS_coding_system) for the geographies that you
	// want data for. Can be:
	// - single values (e.g. E01000001)
	// - comma-separated array of values (e.g E01000001,E01000002,E01000003)
	// - ellipsis-separated contiguous range of values (e.g. E01000001...E01000010)
	//
	// Multiple rows parameters can be supplied, e.g. rows=E01000001&rows=E01000001...E01000010
	Rows *[]string `json:"rows,omitempty"`

	// The census data that you want (NB - use metadata endpoint to see list of currently available census data). Can be:
	// - single values (e.g. QS101EW0001)
	// - comma-separated array of values (e.g QS101EW0001,QS101EW0002,QS101EW0003)
	// - ellipsis-separated contiguous range of values (e.g. QS101EW0001...QS101EW0010)
	//
	// Multiple cols parameters can be supplied, e.g. cols=QS101EW0001&rows=QS101EW0001...QS101EW0010
	Cols *[]string `json:"cols,omitempty"`

	// Two long, lat coordinate pairs representing the opposite corners of a bounding box (e.g. bbox=0.1338,51.4635,0.1017,51.4647).
	// This will select all geographies that lie within this bounding box. Bbox can be used instead of, or in combination with the
	// rows parameter as a way of selecting geography.
	Bbox *string `json:"bbox,omitempty"`

	// Geotype filters API results to a specific geography type. Can be single values or comma-separated array.
	// At the moment these options are supported:
	// - LAD
	// - LSOA
	//
	// Multiple geotype parameters can be supplied, e.g. geotype=LAD&geotype=LSOA
	Geotype *[]string `json:"geotype,omitempty"`

	// Radius and location (both are required) will select all geographies with radius of the long,lat pair location,
	// e.g. location=0.1338,51.4635&radius=1000. Radius and location can be used instead of, or in combination with the rows parameter as a way of selecting geography.
	Location *string `json:"location,omitempty"`

	// Radius and location (both are required) will select all geographies with radius of the long,lat pair location,
	// e.g. location=0.1338,51.4635&radius=1000. Radius and location can be used instead of, or in combination with the rows parameter as a way of selecting geography.
	Radius *int `json:"radius,omitempty"`

	// A sequence of long, lat coordinate pairs representing a closed polygon (NB - 'closed' means the first and last coordinate pair
	// must be the same), e.g. polygon=0.0844,51.4897,0.1214,51.4910,0.1338,51.4635,0.1017,51.4647,0.0844,51.4897. This will select
	// all geographies that lie within this polygon. polygon can be used instead of, or in combination with the rows parameter as a
	// way of selecting geography.
	Polygon     *string `json:"polygon,omitempty"`
	Censustable *string `json:"censustable,omitempty"`

	// (OPTIONAL) - census data category to use as denominator (cat/divide_by).
	// When divide_by is given, the returned value for a category will be the value of the category divided by the value of the
	// divide_by category.
	DivideBy *string `json:"divide_by,omitempty"`
}

// GetQueryParams defines parameters for GetQuery.
type GetQueryParams struct {
	// [ONS codes](https://en.wikipedia.org/wiki/ONS_coding_system) for the geographies that you
	// want data for. Can be:
	// - single values (e.g. E01000001)
	// - comma-separated array of values (e.g E01000001,E01000002,E01000003)
	// - ellipsis-separated contiguous range of values (e.g. E01000001...E01000010)
	//
	// Multiple rows parameters can be supplied, e.g. rows=E01000001&rows=E01000001...E01000010
	Rows *[]string `json:"rows,omitempty"`

	// The census data that you want (NB - use metadata endpoint to see list of currently available census data). Can be:
	// - single values (e.g. QS101EW0001)
	// - comma-separated array of values (e.g QS101EW0001,QS101EW0002,QS101EW0003)
	// - ellipsis-separated contiguous range of values (e.g. QS101EW0001...QS101EW0010)
	//
	// Multiple cols parameters can be supplied, e.g. cols=QS101EW0001&rows=QS101EW0001...QS101EW0010
	Cols *[]string `json:"cols,omitempty"`

	// Two long, lat coordinate pairs representing the opposite corners of a bounding box (e.g. bbox=0.1338,51.4635,0.1017,51.4647).
	// This will select all geographies that lie within this bounding box. Bbox can be used instead of, or in combination with the
	// rows parameter as a way of selecting geography.
	Bbox *string `json:"bbox,omitempty"`

	// Geotype filters API results to a specific geography type. Can be single values or comma-separated array.
	// At the moment these options are supported:
	// - LAD
	// - LSOA
	//
	// Multiple geotype parameters can be supplied, e.g. geotype=LAD&geotype=LSOA
	Geotype *[]string `json:"geotype,omitempty"`

	// Radius and location (both are required) will select all geographies with radius of the long,lat pair location,
	// e.g. location=0.1338,51.4635&radius=1000. Radius and location can be used instead of, or in combination with the rows parameter as a way of selecting geography.
	Location *string `json:"location,omitempty"`

	// Radius and location (both are required) will select all geographies with radius of the long,lat pair location,
	// e.g. location=0.1338,51.4635&radius=1000. Radius and location can be used instead of, or in combination with the rows parameter as a way of selecting geography.
	Radius *int `json:"radius,omitempty"`

	// A sequence of long, lat coordinate pairs representing a closed polygon (NB - 'closed' means the first and last coordinate pair
	// must be the same), e.g. polygon=0.0844,51.4897,0.1214,51.4910,0.1338,51.4635,0.1017,51.4647,0.0844,51.4897. This will select
	// all geographies that lie within this polygon. polygon can be used instead of, or in combination with the rows parameter as a
	// way of selecting geography.
	Polygon     *string `json:"polygon,omitempty"`
	Censustable *string `json:"censustable,omitempty"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// calculate ckmeans over a given category and geography type
	// (GET /ckmeans/{year})
	GetCkmeansYear(w http.ResponseWriter, r *http.Request, year int, params GetCkmeansYearParams)
	// calculate ckmeans for the ratio between two given categories (cat1 / cat2) for a given geography type
	// (GET /ckmeansratio/{year})
	GetCkmeansratioYear(w http.ResponseWriter, r *http.Request, year int, params GetCkmeansratioYearParams)
	// remove all entries from request cache
	// (GET /clear-cache)
	GetClearCache(w http.ResponseWriter, r *http.Request)
	// Get geographic info about an area. Queryable with either geocode or geoname (but not both)
	// (GET /geo/{year})
	GetGeo(w http.ResponseWriter, r *http.Request, year int, params GetGeoParams)
	// Get Metadata
	// (GET /metadata/{year})
	GetMetadataYear(w http.ResponseWriter, r *http.Request, year int, params GetMetadataYearParams)
	// return MSOA code and its name
	// (GET /msoa/{postcode})
	GetMsoaPostcode(w http.ResponseWriter, r *http.Request, postcode string)
	// query census
	// (GET /query/{year})
	GetQueryYear(w http.ResponseWriter, r *http.Request, year int, params GetQueryYearParams)
	// List geocodes matching search conditions
	// (GET /query2/{year})
	GetQuery(w http.ResponseWriter, r *http.Request, year int, params GetQueryParams)
	// spec
	// (GET /swagger)
	GetSwagger(w http.ResponseWriter, r *http.Request)
	// spec
	// (GET /swaggerui)
	GetSwaggerui(w http.ResponseWriter, r *http.Request)
	// CORS preflight OPTIONS request
	// (OPTIONS /{path}/{year})
	Preflight(w http.ResponseWriter, r *http.Request, path string, year int)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// GetCkmeansYear operation middleware
func (siw *ServerInterfaceWrapper) GetCkmeansYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetCkmeansYearParams

	// ------------- Optional query parameter "cat" -------------
	if paramValue := r.URL.Query().Get("cat"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cat", r.URL.Query(), &params.Cat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter cat: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "geotype" -------------
	if paramValue := r.URL.Query().Get("geotype"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geotype", r.URL.Query(), &params.Geotype)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geotype: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "k" -------------
	if paramValue := r.URL.Query().Get("k"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "k", r.URL.Query(), &params.K)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter k: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "divide_by" -------------
	if paramValue := r.URL.Query().Get("divide_by"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "divide_by", r.URL.Query(), &params.DivideBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter divide_by: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetCkmeansYear(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetCkmeansratioYear operation middleware
func (siw *ServerInterfaceWrapper) GetCkmeansratioYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetCkmeansratioYearParams

	// ------------- Optional query parameter "cat1" -------------
	if paramValue := r.URL.Query().Get("cat1"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cat1", r.URL.Query(), &params.Cat1)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter cat1: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "cat2" -------------
	if paramValue := r.URL.Query().Get("cat2"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cat2", r.URL.Query(), &params.Cat2)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter cat2: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "geotype" -------------
	if paramValue := r.URL.Query().Get("geotype"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geotype", r.URL.Query(), &params.Geotype)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geotype: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "k" -------------
	if paramValue := r.URL.Query().Get("k"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "k", r.URL.Query(), &params.K)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter k: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetCkmeansratioYear(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetClearCache operation middleware
func (siw *ServerInterfaceWrapper) GetClearCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetClearCache(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetGeo operation middleware
func (siw *ServerInterfaceWrapper) GetGeo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetGeoParams

	// ------------- Optional query parameter "geocode" -------------
	if paramValue := r.URL.Query().Get("geocode"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geocode", r.URL.Query(), &params.Geocode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geocode: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "geoname" -------------
	if paramValue := r.URL.Query().Get("geoname"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geoname", r.URL.Query(), &params.Geoname)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geoname: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetGeo(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetMetadataYear operation middleware
func (siw *ServerInterfaceWrapper) GetMetadataYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetMetadataYearParams

	// ------------- Optional query parameter "filtertotals" -------------
	if paramValue := r.URL.Query().Get("filtertotals"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "filtertotals", r.URL.Query(), &params.Filtertotals)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter filtertotals: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetMetadataYear(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetMsoaPostcode operation middleware
func (siw *ServerInterfaceWrapper) GetMsoaPostcode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "postcode" -------------
	var postcode string

	err = runtime.BindStyledParameter("simple", false, "postcode", chi.URLParam(r, "postcode"), &postcode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter postcode: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetMsoaPostcode(w, r, postcode)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetQueryYear operation middleware
func (siw *ServerInterfaceWrapper) GetQueryYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetQueryYearParams

	// ------------- Optional query parameter "rows" -------------
	if paramValue := r.URL.Query().Get("rows"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "rows", r.URL.Query(), &params.Rows)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter rows: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "cols" -------------
	if paramValue := r.URL.Query().Get("cols"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cols", r.URL.Query(), &params.Cols)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter cols: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "bbox" -------------
	if paramValue := r.URL.Query().Get("bbox"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "bbox", r.URL.Query(), &params.Bbox)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter bbox: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "geotype" -------------
	if paramValue := r.URL.Query().Get("geotype"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geotype", r.URL.Query(), &params.Geotype)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geotype: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "location" -------------
	if paramValue := r.URL.Query().Get("location"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "location", r.URL.Query(), &params.Location)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter location: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "radius" -------------
	if paramValue := r.URL.Query().Get("radius"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "radius", r.URL.Query(), &params.Radius)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter radius: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "polygon" -------------
	if paramValue := r.URL.Query().Get("polygon"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "polygon", r.URL.Query(), &params.Polygon)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter polygon: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "censustable" -------------
	if paramValue := r.URL.Query().Get("censustable"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "censustable", r.URL.Query(), &params.Censustable)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter censustable: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "divide_by" -------------
	if paramValue := r.URL.Query().Get("divide_by"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "divide_by", r.URL.Query(), &params.DivideBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter divide_by: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetQueryYear(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetQuery operation middleware
func (siw *ServerInterfaceWrapper) GetQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetQueryParams

	// ------------- Optional query parameter "rows" -------------
	if paramValue := r.URL.Query().Get("rows"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "rows", r.URL.Query(), &params.Rows)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter rows: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "cols" -------------
	if paramValue := r.URL.Query().Get("cols"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "cols", r.URL.Query(), &params.Cols)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter cols: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "bbox" -------------
	if paramValue := r.URL.Query().Get("bbox"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "bbox", r.URL.Query(), &params.Bbox)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter bbox: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "geotype" -------------
	if paramValue := r.URL.Query().Get("geotype"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "geotype", r.URL.Query(), &params.Geotype)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter geotype: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "location" -------------
	if paramValue := r.URL.Query().Get("location"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "location", r.URL.Query(), &params.Location)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter location: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "radius" -------------
	if paramValue := r.URL.Query().Get("radius"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "radius", r.URL.Query(), &params.Radius)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter radius: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "polygon" -------------
	if paramValue := r.URL.Query().Get("polygon"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "polygon", r.URL.Query(), &params.Polygon)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter polygon: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "censustable" -------------
	if paramValue := r.URL.Query().Get("censustable"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "censustable", r.URL.Query(), &params.Censustable)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter censustable: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetQuery(w, r, year, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetSwagger operation middleware
func (siw *ServerInterfaceWrapper) GetSwagger(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetSwagger(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetSwaggerui operation middleware
func (siw *ServerInterfaceWrapper) GetSwaggerui(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetSwaggerui(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// Preflight operation middleware
func (siw *ServerInterfaceWrapper) Preflight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "path" -------------
	var path string

	err = runtime.BindStyledParameter("simple", false, "path", chi.URLParam(r, "path"), &path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Path parameter "year" -------------
	var year int

	err = runtime.BindStyledParameter("simple", false, "year", chi.URLParam(r, "year"), &year)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter year: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Preflight(w, r, path, year)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL     string
	BaseRouter  chi.Router
	Middlewares []MiddlewareFunc
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/ckmeans/{year}", wrapper.GetCkmeansYear)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/ckmeansratio/{year}", wrapper.GetCkmeansratioYear)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/clear-cache", wrapper.GetClearCache)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/geo/{year}", wrapper.GetGeo)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/metadata/{year}", wrapper.GetMetadataYear)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/msoa/{postcode}", wrapper.GetMsoaPostcode)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/query/{year}", wrapper.GetQueryYear)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/query2/{year}", wrapper.GetQuery)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/swagger", wrapper.GetSwagger)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/swaggerui", wrapper.GetSwaggerui)
	})
	r.Group(func(r chi.Router) {
		r.Options(options.BaseURL+"/{path}/{year}", wrapper.Preflight)
	})

	return r
}
