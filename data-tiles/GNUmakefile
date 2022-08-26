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

.DEFAULT_GOAL: all

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

# S3 prefix for downloading raw geojson files (in dp-sandbox environment)
S3_URL=s3://ons-dp-sandbox-atlas-input/geojson

# Raw LAD geojson
RAW_LAD=Local_Authority_Districts_(December_2017)_Boundaries_in_the_UK_(WGS84).geojson

# Raw LSOA geojson
RAW_LSOA=Lower_Layer_Super_Output_Areas_(December_2011)_Boundaries_Super_Generalised_Clipped_(BSC)_EW_V3.geojson

# Raw MSOA geojson
RAW_MSOA=Middle_Layer_Super_Output_Areas_(December_2011)_Boundaries_Super_Generalised_Clipped_(BSC)_EW_V3.geojson

# Raw OA geojson
RAW_OA=Output_Areas__December_2011__Boundaries_EW_BGC.geojson

# Name of CSV holding MSOA names
MSOA_NAMES=MSOA-Names-1.16.csv

# URL to download MSOA names
MSOA_URL=https://houseofcommonslibrary.github.io/msoanames/$(MSOA_NAMES)

#
# Directory locations
#
D=data
DD=$D/downloads
DDG=$(DD)/geo
DDM=$(DD)/metrics
DP=$D/processed
DPG=$(DP)/geo
DPM=$(DP)/metrics
DO=$D/output
DOG=$(DO)/geo
DOB=$(DO)/breaks
DOT=$(DO)/tiles


# All the raw geojson files to download
RAW_GEOJSON=$(RAW_LAD) $(RAW_LSOA) $(RAW_MSOA) $(RAW_OA)

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
	mkdir -p $(DDG) $(DDM) $(DPG) $(DPM) $(DOG) $(DOB) $(DOT)

#
# binaries
#
BINARIES=\
	generate-breaks \
	generate-tiles \
	normalise \
	recode-lads \
	rename-msoas \
	split-geojson \
	split-metrics

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

# How to download .geojson files
%.geojson:
	aws s3 cp $(S3_URL)/`basename "$@"` "$@".tmp
	mv "$@".tmp "$@"

# paths to local downloaded zip files
ZIP_PATHS=$(foreach zip,$(NOMIS_ZIPS),$(DDM)/$(zip))

# paths to local downloaded raw geojson files
GEOJSON_PATHS=$(foreach geojson,$(RAW_GEOJSON),$(DDG)/$(geojson))

# path to local downloaded MSOA name file
MSOA_NAME_PATH=data/downloads/geo/$(MSOA_NAMES)

.PHONY: download
download: $(ZIP_PATHS) $(GEOJSON_PATHS) $(MSOA_NAME_PATH)
realclean::
	rm -f $(foreach path,$(GEOJSON_PATHS),"$(path)")
	rm -f $(foreach path,$(ZIP_PATHS),"$(path)")

$(MSOA_NAME_PATH):
	./atomic.sh "$(MSOA_NAME_PATH)" curl "$(MSOA_URL)"
realclean::
	rm -f "$(MSOA_NAME_PATH)"

#
# extract
#

# Paths to extract directories.
# Zip files are extracted into a directory with an .extracted suffix.
EXTRACT_PATHS=$(patsubst %.zip,%.extracted,$(ZIP_PATHS))

# how to extract a zip file
$(EXTRACT_PATHS): %.extracted: %.zip
	./atomic-rm.sh "$@"
	unzip -d "$@" "$<"

.PHONY: extract
extract: $(EXTRACT_PATHS)
clean::
	./atomic-rm.sh $(foreach path,$(EXTRACT_PATHS),"$(path)")

#
# standard
#
# The files in $DPG are versions of the downloaded geojson, but with bboxes added
# for each feature, and with geotype, geocode, ename and wname properties added.
#
# Also MSOA names are added, and certain LAD names are changed.

STANDARD_LAD=$(DPG)/lad.geojson

STANDARD_LSOA=$(DPG)/lsoa.geojson

STANDARD_MSOA=$(DPG)/msoa.geojson

STANDARD_OA=$(DPG)/oa.geojson

.PHONY: standard
standard: $(STANDARD_LAD) $(STANDARD_LSOA) $(STANDARD_MSOA) $(STANDARD_OA)

$(STANDARD_LAD): $(DDG)/$(RAW_LAD) recode-lads.csv recode-lads normalise
	./atomic.sh "$@" sh -o pipefail -c ' \
		./recode-lads -r recode-lads.csv < "$(DDG)/$(RAW_LAD)" | \
		./normalise -t LAD -c lad17cd -e lad17nm -w lad17nmw \
	'
clean::
	rm -f "$(STANDARD_LAD)"

$(STANDARD_LSOA): $(DDG)/$(RAW_LSOA) ./normalise
	./atomic.sh "$@" sh -o pipefail -c ' \
		./normalise -t LSOA -c LSOA11CD -e LSOA11NM -w LSOA11NMW < "$(DDG)/$(RAW_LSOA)" \
	'
clean::
	rm -f "$(STANDARD_LSOA)"


$(STANDARD_MSOA): $(DDG)/$(RAW_MSOA) $(DDG)/$(MSOA_NAMES) rename-msoas normalise
	./atomic.sh "$@" sh -o pipefail -c ' \
		./rename-msoas -n "$(DDG)/$(MSOA_NAMES)" < "$(DDG)/$(RAW_MSOA)" | \
		./normalise -t MSOA -c MSOA11CD -e MSOA11NM -e MSOA11NMW \
	'
clean::
	rm -f "$(STANDARD_MSOA)"

$(STANDARD_OA): $(DDG)/$(RAW_OA) normalise
	./atomic.sh "$@" sh -o pipefail -c ' \
		./normalise -t OA -c OA11CD -e OA11CD -w OA11CD < "$(DDG)/$(RAW_OA)" \
	'
clean::
	rm -f "$(STANDARD_OA)"

#
# geos
#
# All the output geojson files end up in a a single directory, so it is all or nothing.

.PHONY: geos
geos: $(DOG)/.done

DOG_TMP=$(DOG).tmp

$(DOG)/.done: $(STANDARD_LAD) $(STANDARD_LSOA) $(STANDARD_MSOA) $(STANDARD_OA) split-geojson
	./atomic-rm.sh "$(DOG_TMP)"
	mkdir "$(DOG_TMP)"
	./split-geojson -d "$(DOG_TMP)" < "$(STANDARD_LAD)"
	./split-geojson -d "$(DOG_TMP)" < "$(STANDARD_LSOA)"
	./split-geojson -d "$(DOG_TMP)" < "$(STANDARD_MSOA)"
	./split-geojson -d "$(DOG_TMP)" < "$(STANDARD_OA)"
	touch "$(DOG_TMP)"/.done
	./atomic-rm.sh "$(DOG)"
	mv "$(DOG_TMP)" "$(DOG)"
clean::
	./atomic-rm.sh "$(DOG_TMP)" "$(DOG)"

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
		$(STANDARD_LAD) \
		$(STANDARD_LSOA) \
		$(STANDARD_MSOA) \
		$(STANDARD_OA) \
		$(DPM)/.done \
		content.json \
		categories.txt \
		generate-tiles
	./atomic-rm.sh "$(DOT_TMP)"
	mkdir "$(DOT_TMP)"
	./generate-tiles \
		-G "$(DPG)" \
		-M "$(DPM)" \
		-c categories.txt \
		-q content.json \
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
		$(STANDARD_LAD) \
		$(STANDARD_LSOA) \
		$(STANDARD_MSOA) \
		$(STANDARD_OA) \
		$(DPM)/.done \
		content.json \
		categories.txt \
		generate-breaks
	./atomic-rm.sh "$(DOB_TMP)"
	mkdir "$(DOB_TMP)"
	./generate-breaks \
		-G "$(DPG)" \
		-M "$(DPM)" \
		-c categories.txt \
		-q content.json \
		-O "$(DOB_TMP)"
	touch "$(DOB_TMP)"/.done
	./atomic-rm.sh "$(DOB)"
	mv "$(DOB_TMP)" "$(DOB)"
clean::
	./atomic-rm.sh "$(DOB_TMP)" "$(DOB)"
