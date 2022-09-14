#
# This makefile requires GNU make
#
# Important targets:
#
#	all -- do everything
#
#	dirs -- create data directory structure
#	binaries -- compile Go programs (you may need -B)
#
#	download -- download geojson and metrics files
#	download-cat -- download files for generating cateogry list
#	download-met -- download just metrics
#	download-geo -- download just geojson files
#
#	standard -- normalise downloaded geojson and metrics files
#	standard-cat -- generate categories.txt
#	standard-geo -- normalise just downloaded geojson (add bbox, set geotype, etc)
#	standard-met -- normalise just downloaded metrics files to standard CSVs
#
#	single-metrics -- split standard metrics CSVs into single-metric files
#
#	geos -- generate geojson files for every geography in standard geojson
#	tiles - generate data tile CSVs
#	breaks -- generate break
#
#	clean -- remove generated files
#	realclean -- remove generated and downloaded files
#
# Command line variable overrides
#
#	D -- name of data directory (default "data")
#	DATA_TILE_GRID -- name of quads and layers file (default "DataTileGrid.json")
#	CONTENT -- name of content.json file (default "content.json")
#	GEOVERSION -- geography version (default "2011")
#	METVERSION -- metrics version (default "2011")
#	CATVERSION -- content.json version (default "2011")
#
# Or you can create a params.mk file and set those variables.
#
# Or you can say PARAMS=some-params.mk

.DEFAULT_GOAL=all
SHELL=bash

# Try to include params.mk or $PARAMS, but no error if it's missing.
PARAMS?=params.mk
-include $(PARAMS)

#
# Config files
#
DATA_TILE_GRID?=DataTileGrid.json
CONTENT?=content.json

GEOVERSION?=2011
METVERSION?=2011
CATVERSION?=2011

#
# Directory locations
#
D?=data

DD=$D/downloads
DDG=$(DD)/geo
DDGV=$(DDG)/$(GEOVERSION)
DDM=$(DD)/metrics
DDMV=$(DDM)/$(METVERSION)

DP=$D/processed
DPG=$(DP)/geo
DPGV=$(DPG)/$(GEOVERSION)
DPM=$(DP)/metrics
DPMV=$(DPM)/$(METVERSION)

DO=$D/output/geo-$(GEOVERSION)+met-$(METVERSION)
DOG=$(DO)/geo
DOB=$(DO)/breaks
DOT=$(DO)/tiles

include geo-$(GEOVERSION).mk
include met-$(METVERSION).mk
include cat-$(CATVERSION).mk

#
# do everything
#
.PHONY: all
all: dirs $(DOG)/.done $(DOT)/.done $(DOB)/.done

#
# cleanup
#
.PHONY: realclean
realclean:: clean
.PHONY: clean
clean::

#
# data directory structure
#
.PHONY: dirs
dirs:
	mkdir -p $(DDGV) $(DDMV) $(DPGV) $(DPMV) $(DOG) $(DOB) $(DOT)

#
# binaries
#
BINARIES=\
	extract-categories \
	fake-data \
	generate-breaks \
	generate-tiles \
	map-categories \
	normalise \
	recode-lads \
	rename-msoas \
	split-geojson \
	split-metrics \
	xls2csv

.PHONY: binaries
binaries: $(BINARIES)

$(BINARIES):
	GOBIN=`pwd` go install ./cmd/$@
clean::
	rm -f $(BINARIES)

#
# download targets
#

.PHONY: download-cat download-met download-geo download
download-cat: $(CAT_DOWNLOADS)
download-geo: $(GEO_DOWNLOADS)
download-met: $(MET_DOWNLOADS)
download: $(CAT_DOWNLOADS) $(GEO_DOWNLOADS) $(MET_DOWNLOADS)

#
# processing and normalising
#

.PHONY: standard-cat standard-geo standard-met standard
standard-cat: $(CAT_STANDARD)
standard-geo: $(GEO_STANDARD)
standard-met: $(MET_STANDARD)
standard: $(GEO_STANDARD) $(MET_STANDARD)

#
# geos
#
# All the output geojson files end up in a a single directory, so it is all or nothing.

.PHONY: geos
geos: $(DOG)/.done

DOG_TMP=$(DOG).tmp

$(DOG)/.done: $(GEO_STANDARD) split-geojson
	./atomic-rm.sh "$(DOG_TMP)"
	mkdir "$(DOG_TMP)"
	for f in $(foreach processed,$(GEO_STANDARD),"$(processed)") ;\
	do \
		echo "splitting $$f" ;\
		./split-geojson -d "$(DOG_TMP)" < "$$f" ;\
	done
	touch "$(DOG_TMP)"/.done
	./atomic-rm.sh "$(DOG)"
	mv "$(DOG_TMP)" "$(DOG)"
clean::
	./atomic-rm.sh "$(DOG_TMP)" "$(DOG)"

#
# metrics
#
# All the single-metric files end up in the same directory, so it is all or nothing.

.PHONY: single-metrics
single-metrics: $(DPMV)/.done

DPMV_TMP=$(DPMV).tmp

# XXX may need to change split-metrics to take list of directories and/or different
# XXX glob name other than *DATA.CSV
$(DPMV)/.done: $(MET_STANDARD) split-metrics
	./atomic-rm.sh "$(DPMV_TMP)"
	mkdir "$(DPMV_TMP)"
	./split-metrics -s "$(DDMV)" -d "$(DPMV_TMP)"
	touch "$(DPMV_TMP)"/.done
	./atomic-rm.sh "$(DPMV)"
	mv "$(DPMV_TMP)" "$(DPMV)"
clean::
	./atomic-rm.sh "$(DPMV_TMP)" "$(DPMV)"

#
# tiles
#

.PHONY: tiles
tiles: $(DOT)/.done

DOT_TMP=$(DOT).tmp

$(DOT)/.done: \
		$(GEO_STANDARD) \
		$(DPMV)/.done \
		$(DATA_TILE_GRID) \
		$(CAT_STANDARD) \
		generate-tiles
	./atomic-rm.sh "$(DOT_TMP)"
	mkdir "$(DOT_TMP)"
	./generate-tiles \
		-G "$(DPGV)" \
		-M "$(DPMV)" \
		-c "$(CAT_STANDARD)" \
		-q "$(DATA_TILE_GRID)" \
		-O "$(DOT_TMP)"
	touch "$(DOT_TMP)"/.done
	./atomic-rm.sh "$(DOT)"
	mv "$(DOT_TMP)" "$(DOT)"
clean::
	./atomic-rm.sh "$(DOT_TMP)" "$(DOT)"

#
# breaks
#

.PHONY: breaks
breaks: $(DOB)/.done

DOB_TMP=$(DOB).tmp

$(DOB)/.done: \
		$(GEO_STANDARD) \
		$(DPMV)/.done \
		$(DATA_TILE_GRID) \
		$(CAT_STANDARD) \
		generate-breaks
	./atomic-rm.sh "$(DOB_TMP)"
	mkdir "$(DOB_TMP)"
	./generate-breaks \
		-G "$(DPGV)" \
		-M "$(DPMV)" \
		-c "$(CAT_STANDARD)" \
		-q "$(DATA_TILE_GRID)" \
		-O "$(DOB_TMP)"
	touch "$(DOB_TMP)"/.done
	./atomic-rm.sh "$(DOB)"
	mv "$(DOB_TMP)" "$(DOB)"
clean::
	./atomic-rm.sh "$(DOB_TMP)" "$(DOB)"

#
# tests
#

test:: extract-categories
	diff \
		cmd/extract-categories/testdata/categories.txt \
		<( ./extract-categories < cmd/extract-categories/testdata/content.json )
