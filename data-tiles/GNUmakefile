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
#	download-met -- download just metrics
#	download-geo -- download just geojson files
#
#	process-geo -- normalise downloaded geojson (add bbox, set geotype, etc)
#	process-met -- process metrics files to standard CSVs
#
#	metrics -- split metrics CSVs into single-metric files
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
#	CATEGORIES -- name of categories file (default "categories.txt")
#	GEOVERSION -- geography version (default "2011")
#	METVERSION -- metrics version (default "2011")

.DEFAULT_GOAL=all

#
# Config files
#
DATA_TILE_GRID?=DataTileGrid.json
CATEGORIES?=categories.txt

GEOVERSION?=2011
METVERSION?=2011

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
DO=$D/output
DOG=$(DO)/geo
DOGV=$(DOG)/$(GEOVERSION)
DOB=$(DO)/breaks
DOT=$(DO)/tiles

include geo-$(GEOVERSION).mk
include met-$(METVERSION).mk

#
# do everything
#
.PHONY: all
all: dirs $(DOGV)/.done $(DOT)/.done $(DOB)/.done

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
	mkdir -p $(DDGV) $(DDMV) $(DPGV) $(DPMV) $(DOGV) $(DOB) $(DOT)

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

.PHONY: download-met download-geo download
download-geo: $(GEO_DOWNLOADS)
download-met: $(MET_DOWNLOADS)
download: $(GEO_DOWNLOADS) $(MET_DOWNLOADS)

#
# processing and normalising
#

.PHONY: process-geo process-met process
process-geo: $(GEO_PROCESSED)
process-met: $(MET_PROCESSED)
process: $(GEO_PROCESSED) $(MET_PROCESSED)

#
# geos
#
# All the output geojson files end up in a a single directory, so it is all or nothing.

.PHONY: geos
geos: $(DOGV)/.done

DOGV_TMP=$(DOGV).tmp

$(DOGV)/.done: $(GEO_PROCESSED) split-geojson
	./atomic-rm.sh "$(DOGV_TMP)"
	mkdir "$(DOGV_TMP)"
	for f in $(foreach processed,$(GEO_PROCESSED),"$(processed)") ;\
	do \
		echo "splitting $$f" ;\
		./split-geojson -d "$(DOGV_TMP)" < "$$f" ;\
	done
	touch "$(DOGV_TMP)"/.done
	./atomic-rm.sh "$(DOGV)"
	mv "$(DOGV_TMP)" "$(DOGV)"
clean::
	./atomic-rm.sh "$(DOGV_TMP)" "$(DOGV)"

#
# metrics
#
# All the single-metric files end up in the same directory, so it is all or nothing.

.PHONY: split-metrics
split-metrics: $(DPMV)/.done

DPMV_TMP=$(DPMV).tmp

# XXX may need to change split-metrics to take list of directories and/or different
# XXX glob name other than *DATA.CSV
$(DPMV)/.done: $(MET_PROCESSED) split-metrics
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
		$(GEO_PROCESSED) \
		$(DPM)/.done \
		$(DATA_TILE_GRID) \
		$(CATEGORIES) \
		generate-tiles
	./atomic-rm.sh "$(DOT_TMP)"
	mkdir "$(DOT_TMP)"
	./generate-tiles \
		-G "$(DPGV)" \
		-M "$(DPM)" \
		-c "$(CATEGORIES)" \
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
		$(GEO_PROCESSED) \
		$(DPM)/.done \
		$(DATA_TILE_GRID) \
		$(CATEGORIES) \
		generate-breaks
	./atomic-rm.sh "$(DOB_TMP)"
	mkdir "$(DOB_TMP)"
	./generate-breaks \
		-G "$(DPGV)" \
		-M "$(DPM)" \
		-c "$(CATEGORIES)" \
		-q "$(DATA_TILE_GRID)" \
		-O "$(DOB_TMP)"
	touch "$(DOB_TMP)"/.done
	./atomic-rm.sh "$(DOB)"
	mv "$(DOB_TMP)" "$(DOB)"
clean::
	./atomic-rm.sh "$(DOB_TMP)" "$(DOB)"
