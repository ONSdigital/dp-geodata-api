#
# This makefile requires GNU make
#
# Run "make help" to get help about important targets and variables to set.
#
# You can set variables on the make command line, or set them in a params.mk file.
# Use PARAMS=some-params.mk to use a specific params.mk file:
#
#	make all PARAMS=some-params.mk
#

.DEFAULT_GOAL=help
SHELL=bash

# Try to include params.mk or $PARAMS, but no error if it's missing.
#= name of optional parameters file
PARAMS?=params.mk
-include $(PARAMS)

#
# Default datasets to use
#
#= name of quads file
DATA_TILE_GRID?=DataTileGrid.json
#= version of geography data to use
GEOVERSION?=2011
#= version of metrics data to use
METVERSION?=2011
#= version of categories.txt/content.json to use
CATVERSION?=2011
#= set to "-R" to calculate metric ratios
CALC_RATIOS?=

#
# Directory locations
#
#= location of output directory
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

#= output directory
DO?=$D/output
DOG=$(DO)/geo
DOB=$(DO)/breaks
DOT=$(DO)/tiles

#
# Include snippets describing how to grab/process specific datasets.
# Be careful about variables and order here. At present, met-fake has
# to come after cat and geo because it needs to know which geos and
# categories to create fake metrics for.
#
include geo-$(GEOVERSION).mk
include cat-$(CATVERSION).mk
include met-$(METVERSION).mk

#
# help
#
# (We have to use join because dependencies may be split over multiple lines.)
.PHONY: help
help:	## this help
	@echo "Important Targets"
	@./join.sh $(MAKEFILE_LIST) | egrep -h '\s##\s' | sort | awk 'BEGIN{FS=":.*?## "}{printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo "Important Variables (set on command line)"
	@./join2.sh $(MAKEFILE_LIST) | egrep -h '^#=' | awk 'BEGIN{FS="#=|\\??="}{printf "\033[36m%-20s\033[0m%s (\033[36m%s\033[0m)\n", $$3, $$2, $$4}' | sort

#
# do everything
#
.PHONY: all
all: dirs $(DOG)/.done $(DOT)/.done $(DOB)/.done	## build all output files

#
# cleanup
#
.PHONY: realclean
realclean:: clean	## remove downloaded and generated files
.PHONY: clean
clean::		## remove generated files

#
# data directory structure
#
.PHONY: dirs
dirs:	## create output directories (must set Important Variables)
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
binaries: $(BINARIES)	## build Go binaries (probably want -B also)

$(BINARIES):
	GOBIN=`pwd` go install ./cmd/$@
clean::
	rm -f $(BINARIES)

#
# download targets
#

.PHONY: download-cat download-met download-geo download
download-cat: $(CAT_DOWNLOADS)	## download content.json/categories.txt
download-geo: $(GEO_DOWNLOADS)	## download geojson and related files
download-met: $(MET_DOWNLOADS)	## download metrics (eg census CSVs)
download: $(CAT_DOWNLOADS) $(GEO_DOWNLOADS) $(MET_DOWNLOADS)	## download everything

#
# processing and normalising
#

.PHONY: standard-cat standard-geo standard-met standard
standard-cat: $(CAT_STANDARD)	## normalise content.json/categories.txt
standard-geo: $(GEO_STANDARD)	## normalise geojson files
standard-met: $(MET_STANDARD)	## normalise metrics CSVs
standard: $(GEO_STANDARD) $(MET_STANDARD)	## normalise everything

#
# geos
#
# All the output geojson files end up in a a single directory, so it is all or nothing.

.PHONY: geos
geos: $(DOG)/.done	## generate output geojson files

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
single-metrics: $(DPMV)/.done	## generate output single-metrics CSVs

DPMV_TMP=$(DPMV).tmp

$(DPMV)/.done: $(MET_STANDARD) split-metrics
	./atomic-rm.sh "$(DPMV_TMP)"
	mkdir "$(DPMV_TMP)"
	./split-metrics $(CALC_RATIOS) -s "$(DDMV)" -d "$(DPMV_TMP)"
	touch "$(DPMV_TMP)"/.done
	./atomic-rm.sh "$(DPMV)"
	mv "$(DPMV_TMP)" "$(DPMV)"
clean::
	./atomic-rm.sh "$(DPMV_TMP)" "$(DPMV)"

#
# tiles
#

.PHONY: tiles
tiles: $(DOT)/.done	## generate output data tile files

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
breaks: $(DOB)/.done	## generate output breaks files

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
# tests -- some rough cli sanity tests
#

# If a test writes output, use TEST_OUTPUT directory.
TEST_OUTPUT=.test-output
clean::
	./atomic-rm.sh ./test-output

.PHONY: test
test: \
	test-extract-categories \
	test-fake-data \
	test-generate-breaks \
	test-generate-tiles \
	test-map-categories \
	test-normalise	\
	test-recode-lads \
	test-rename-msoas \
	test-split-geojson \
	test-split-metrics \
	test-xls2csv	## run all tests

.PHONY: test-extract-categories
test-extract-categories: extract-categories	## test extract-categories cli
	diff \
		cmd/extract-categories/testdata/old-categories.txt \
		<( ./extract-categories < cmd/extract-categories/testdata/old-content.json )
	diff \
		cmd/extract-categories/testdata/new-categories.txt \
		<( ./extract-categories < cmd/extract-categories/testdata/new-content.json )


.PHONY: test-fake-data
test-fake-data: fake-data	## test fake-data cli
	diff \
		cmd/fake-data/testdata/DATA.CSV \
		<( ./fake-data \
			-G cmd/fake-data/testdata \
			-c cmd/fake-data/testdata/categories.txt \
			-r 0 \
		)

.PHONY: test-generate-breaks
test-generate-breaks: \
	test-generate-breaks-ratio \
	test-generate-breaks-non-ratio	## test generate-breaks

.PHONY: test-generate-breaks-ratio
test-generate-breaks-ratio: generate-breaks	## test generate-breaks cli in ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./generate-breaks \
		-R \
		-G cmd/generate-breaks/testdata/geo \
		-M cmd/generate-breaks/testdata/metrics \
		-O "$(TEST_OUTPUT)" \
		-c cmd/generate-breaks/testdata/categories.txt \
		-q cmd/generate-breaks/testdata/quads.json
	diff -r cmd/generate-breaks/testdata/output-ratio "$(TEST_OUTPUT)"

.PHONY: test-generate-breaks-non-ratio
test-generate-breaks-non-ratio: generate-breaks	## test generate-breaks cli in non-ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./generate-breaks \
		-G cmd/generate-breaks/testdata/geo \
		-M cmd/generate-breaks/testdata/metrics \
		-O "$(TEST_OUTPUT)" \
		-c cmd/generate-breaks/testdata/categories.txt \
		-q cmd/generate-breaks/testdata/quads.json
	diff -r cmd/generate-breaks/testdata/output-non-ratio "$(TEST_OUTPUT)"


.PHONY: test-generate-tiles
test-generate-tiles: \
	test-generate-tiles-ratio \
	test-generate-tiles-non-ratio	## test generate-tiles cli

test-generate-tiles-ratio: generate-tiles	## test generate-tiles cli in ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./generate-tiles \
		-R \
		-G cmd/generate-tiles/testdata/geo \
		-M cmd/generate-tiles/testdata/metrics \
		-O "$(TEST_OUTPUT)" \
		-c cmd/generate-tiles/testdata/categories.txt \
		-q cmd/generate-tiles/testdata/quads.json
	diff -r cmd/generate-tiles/testdata/output-ratio "$(TEST_OUTPUT)"

test-generate-tiles-non-ratio: generate-tiles	## test generate-tiles cli in non-ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./generate-tiles \
		-G cmd/generate-tiles/testdata/geo \
		-M cmd/generate-tiles/testdata/metrics \
		-O "$(TEST_OUTPUT)" \
		-c cmd/generate-tiles/testdata/categories.txt \
		-q cmd/generate-tiles/testdata/quads.json
	diff -r cmd/generate-tiles/testdata/output-non-ratio "$(TEST_OUTPUT)"

.PHONY: test-map-categories
test-map-categories: map-categories	## test map-categories cli
	diff \
		cmd/map-categories/testdata/out.csv \
		<( ./map-categories \
			-c cmd/map-categories/testdata/content.json \
			-C legal_partnership_status_6a \
			<  cmd/map-categories/testdata/in.csv \
		)

.PHONY: test-normalise
test-normalise: normalise	## test normalise cli
	diff \
		<( jq < cmd/normalise/testdata/out.geojson ) \
		<( ./normalise \
			-c lad17cd \
			-t LAD \
			-e lad17nm \
			-w lad17nmw \
			< cmd/normalise/testdata/in.geojson | jq \
		)

.PHONY: test-recode-lads
test-recode-lads: recode-lads	## test recode-lads cli
	diff \
		<( jq -S < cmd/recode-lads/testdata/out.geojson ) \
		<( ./recode-lads \
			-r cmd/recode-lads/testdata/recode-lads.csv \
			< cmd/recode-lads/testdata/in.geojson | jq -S \
		)

.PHONY: test-rename-msoas
test-rename-msoas: rename-msoas	## test rename-msoas cli
	diff \
		<( jq -S < cmd/rename-msoas/testdata/out.geojson ) \
		<( ./rename-msoas \
			-n cmd/rename-msoas/testdata/msoa-names.csv \
			< cmd/rename-msoas/testdata/in.geojson | jq -S \
		)

.PHONY: test-split-geojson
test-split-geojson: split-geojson	## test split-geojson cli
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./split-geojson -d "$(TEST_OUTPUT)" < cmd/split-geojson/testdata/in.geojson
	diff -r cmd/split-geojson/testdata/out "$(TEST_OUTPUT)"

.PHONY: test-split-metrics
test-split-metrics: \
	test-split-metrics-non-ratio \
	test-split-metrics-ratio	## test split-metrics cli

test-split-metrics-non-ratio: split-metrics	## test split-metrics cli in non-ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./split-metrics -s cmd/split-metrics/testdata/in -d "$(TEST_OUTPUT)"
	diff -r cmd/split-metrics/testdata/out-non-ratio "$(TEST_OUTPUT)"

test-split-metrics-ratio: split-metrics	## test split-metrics cli in ratio mode
	./atomic-rm.sh "$(TEST_OUTPUT)"
	mkdir "$(TEST_OUTPUT)"
	./split-metrics -R -s cmd/split-metrics/testdata/in -d "$(TEST_OUTPUT)"
	diff -r cmd/split-metrics/testdata/out-ratio "$(TEST_OUTPUT)"

.PHONY: test-xls2csv
test-xls2csv: xls2csv	## test xls2csv cli
	diff \
		cmd/xls2csv/testdata/out.csv \
		<( ./xls2csv < cmd/xls2csv/testdata/in.xlsx )
