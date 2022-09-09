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
#	extract -- extract CSVs from metrics zip files
#	standard -- normalise downloaded geojson (add bbox, set geotype, etc)
#
#	geos -- generate geojson files for every geography in standard geojson
#	metrics -- split downloaded CSVs into single-metric files
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

.DEFAULT_GOAL=all

#
# Config files
#
DATA_TILE_GRID?=DataTileGrid.json
CATEGORIES?=categories.txt

GEOVERSION?=2011

#
# Files to download
#

# Location of nomis zips
NOMIS_URL=https://www.nomisweb.co.uk/output/census/2011

# List of nomis zips to download, taken from Viv's v4 google spreadsheet.
NOMIS_ZIPS=\
	ks103ew_2011_oa.zip \
	ks202ew_2011_oa.zip \
	ks206ew_2011_oa.zip \
	ks207wa_2011_oa.zip \
	ks608ew_2011_oa.zip \
	qs101ew_2011_oa.zip \
	qs103ew_2011_oa.zip \
	qs104ew_2011_oa.zip \
	qs113ew_2011_oa.zip \
	qs119ew_2011_oa.zip \
	qs201ew_2011_oa.zip \
	qs202ew_2011_oa.zip \
	qs203ew_2011_oa.zip \
	qs208ew_2011_oa.zip \
	qs301ew_2011_oa.zip \
	qs302ew_2011_oa.zip \
	qs303ew_2011_oa.zip \
	qs402ew_2011_oa.zip \
	qs403ew_2011_oa.zip \
	qs406ew_2011_oa.zip \
	qs411ew_2011_oa.zip \
	qs415ew_2011_oa.zip \
	qs416ew_2011_oa.zip \
	qs501ew_2011_oa.zip \
	qs601ew_2011_oa.zip \
	qs604ew_2011_oa.zip \
	qs605ew_2011_oa.zip \
	qs701ew_2011_oa.zip \
	qs702ew_oa.zip \
	qs803ew_2011_oa.zip \

#
# Directory locations
#
D?=data
DD=$D/downloads
DDG=$(DD)/geo
DDGV=$(DDG)/$(GEOVERSION)
DDM=$(DD)/metrics
DP=$D/processed
DPG=$(DP)/geo
DPGV=$(DPG)/$(GEOVERSION)
DPM=$(DP)/metrics
DO=$D/output
DOG=$(DO)/geo
DOGV=$(DOG)/$(GEOVERSION)
DOB=$(DO)/breaks
DOT=$(DO)/tiles

include geo-$(GEOVERSION).mk

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
	mkdir -p $(DDGV) $(DDM) $(DPGV) $(DPM) $(DOGV) $(DOB) $(DOT)

#
# binaries
#
BINARIES=\
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
# download
#

# How to download .zip files
%.zip:
	./atomic.sh "$@" curl $(NOMIS_URL)/`basename "$@"`

# paths to local downloaded zip files
ZIP_PATHS=$(foreach zip,$(NOMIS_ZIPS),$(DDM)/$(zip))

.PHONY: download
download: $(ZIP_PATHS) $(GEO_DOWNLOADS)
realclean::
	rm -f $(foreach path,$(ZIP_PATHS),"$(path)")

#
# extract
#

# Paths to extract directories.
# Zip files are extracted into a directory with an .extracted suffix.
EXTRACT_PATHS=$(patsubst %.zip,%.extracted,$(ZIP_PATHS))

# how to extract a zip file
$(EXTRACT_PATHS): %.extracted: %.zip
	./atomic-rm.sh "$@".tmp
	unzip -d "$@".tmp "$<"
	mv "$@".tmp "$@"

.PHONY: extract
extract: $(EXTRACT_PATHS)
clean::
	./atomic-rm.sh $(foreach path,$(EXTRACT_PATHS),"$(path)")

#
# standard
#

.PHONY: standard
standard: $(GEO_PROCESSED)

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

.PHONY: metrics
metrics: $(DPM)/.done

DPM_TMP=$(DPM).tmp

# The single-metric files really depend on all *DATA.CSV files found under $DDM.
# But $(EXTRACT_PATHS) work better because previous rules operate on these directories
# and not on the CSV files.
$(DPM)/.done: $(EXTRACT_PATHS) split-metrics
	./atomic-rm.sh "$(DPM_TMP)"
	mkdir "$(DPM_TMP)"
	./split-metrics -s "$(DDM)" -d "$(DPM_TMP)"
	touch "$(DPM_TMP)"/.done
	./atomic-rm.sh "$(DPM)"
	mv "$(DPM_TMP)" "$(DPM)"
clean::
	./atomic-rm.sh "$(DPM_TMP)" "$(DPM)"

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
