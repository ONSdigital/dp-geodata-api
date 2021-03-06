package handlers

import (
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-geodata-api/api"
	"github.com/ONSdigital/dp-geodata-api/sentinel"
)

func (svr *Server) GetCkmeansYear(w http.ResponseWriter, r *http.Request, year int, params api.GetCkmeansYearParams) {
	if !svr.assertAuthorized(w, r) || !svr.assertDatabaseEnabled(w, r) {
		return
	}

	generate := func() ([]byte, error) {
		var cat, geotype []string
		var divideBy string
		var k int
		if params.Cat != nil {
			cat = *params.Cat
		}
		if params.Geotype != nil {
			geotype = *params.Geotype
		}
		if params.K != nil {
			k = *params.K
		}
		if params.DivideBy != nil {
			divideBy = *params.DivideBy
		}
		if cat == nil || geotype == nil || k == 0 {
			return nil, fmt.Errorf("%w: cat, geotype and k required", sentinel.ErrMissingParams)
		}

		ctx := r.Context()
		breaks, err := svr.querygeodata.CKmeans(ctx, year, cat, geotype, k, divideBy)
		if err != nil {
			return nil, err
		}
		return toJSON(breaks)
	}

	svr.respond(w, r, mimeJSON, generate)
}

// !!!! DEPRECATED CKMEANSRATIO TO BE REMOVED WHEN FRONT END REMOVES DEPENDENCY ON IT !!!!
//
func (svr *Server) GetCkmeansratioYear(w http.ResponseWriter, r *http.Request, year int, params api.GetCkmeansratioYearParams) {
	if !svr.assertAuthorized(w, r) || !svr.assertDatabaseEnabled(w, r) {
		return
	}

	generate := func() ([]byte, error) {
		var cat1, cat2, geotype string
		var k int
		if params.Cat1 != nil {
			cat1 = *params.Cat1
		}
		if params.Cat2 != nil {
			cat2 = *params.Cat2
		}
		if params.Geotype != nil {
			geotype = *params.Geotype
		}
		if params.K != nil {
			k = *params.K
		}
		if cat1 == "" || cat2 == "" || geotype == "" || k == 0 {
			return nil, fmt.Errorf("%w: cat1, cat2, geotype and k required", sentinel.ErrMissingParams)
		}

		ctx := r.Context()
		breaks, err := svr.querygeodata.CKmeansRatio(ctx, year, cat1, cat2, geotype, k)
		if err != nil {
			return nil, err
		}
		return toJSON(breaks)
	}

	svr.respond(w, r, mimeJSON, generate)
}
