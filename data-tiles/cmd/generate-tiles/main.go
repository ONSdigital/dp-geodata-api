package main

import (
	"flag"
	"log"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/cat"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/geos"
	"github.com/ONSdigital/dp-geodata-api/data-tiles/grid"
)

func main() {
	catfile := flag.String("c", "categories.txt", "text file holding list of categories to use")
	gridfile := flag.String("q", "DataTileGrid.json", "json tile description file")
	geodir := flag.String("G", "data/processed/geo", "directory holding geojson files for each geotype")
	metdir := flag.String("M", "data/processed/metrics", "directory holding metrics files for each category")
	outdir := flag.String("O", "data/output/tiles", "output directory")
	flag.Parse()

	catlist, err := cat.LoadCategories(*catfile)
	if err != nil {
		log.Fatal(err)
	}

	metrics, err := cat.LoadMetrics(catlist, *metdir)
	if err != nil {
		log.Fatal(err)
	}

	bounds, err := geos.LoadAll(*geodir)
	if err != nil {
		log.Fatal(err)
	}

	quads, err := grid.Load(*gridfile)
	if err != nil {
		log.Fatal(err)
	}

	err = generateTiles(
		catlist,
		quads,
		bounds,
		metrics,
		*outdir,
	)
	if err != nil {
		log.Fatal(err)
	}
}
