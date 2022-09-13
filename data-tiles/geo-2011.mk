#
# geo-2011.mk -- rules to download and process 2011 geographies
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	GEO_DOWNLOADS	list of downloaded geo files
#	GEO_STANDARD	list of normalised geo files
#
# Each file named in those variables should have targets in this file.

#
# geojson files from S3
#

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

# All the raw geojson files to download
RAW_GEOJSON=$(RAW_LAD) $(RAW_LSOA) $(RAW_MSOA) $(RAW_OA)

# How to download .geojson files
%.geojson:
	aws s3 cp $(S3_URL)/`basename "$@"` "$@".tmp
	mv "$@".tmp "$@"

#
# MSOA names file from House of Commons
#

# Name of CSV holding MSOA names
MSOA_NAMES=MSOA-Names-1.16.csv

# URL to download MSOA names
MSOA_URL=https://houseofcommonslibrary.github.io/msoanames/$(MSOA_NAMES)

# path to local downloaded MSOA name file
MSOA_NAME_PATH=$(DDGV)/$(MSOA_NAMES)

$(MSOA_NAME_PATH):
	./atomic.sh "$(MSOA_NAME_PATH)" curl "$(MSOA_URL)"
realclean::
	rm -f "$(MSOA_NAME_PATH)"


#
# Set GEO_DOWNLOADS so parent makefile can use download files as targets.
#

# paths to local downloaded raw geojson files and MSOA names file
GEO_DOWNLOADS=\
	$(foreach geojson,$(RAW_GEOJSON),$(DDGV)/$(geojson)) \
	$(MSOA_NAME_PATH)
realclean::
	rm -f $(foreach geojson,$(RAW_GEOJSON),"$(DDGV)/$(geojson)")

#
# Processed geo files
#

# The files in $DPG are versions of the downloaded geojson, but with bboxes added
# for each feature, and with geotype, geocode, ename and wname properties added.
#
# Also MSOA names are added, and certain LAD names are changed.

STANDARD_LAD=$(DPGV)/lad.geojson

STANDARD_LSOA=$(DPGV)/lsoa.geojson

STANDARD_MSOA=$(DPGV)/msoa.geojson

STANDARD_OA=$(DPGV)/oa.geojson

#
# Set GEO_STANDARD so parent make file can use processed files as targets.
#
GEO_STANDARD=$(STANDARD_LAD) $(STANDARD_LSOA) $(STANDARD_MSOA) $(STANDARD_OA)

#
# Rules to process geo files
#

$(STANDARD_LAD): $(DDGV)/$(RAW_LAD) recode-lads.csv recode-lads normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
		./recode-lads -r recode-lads.csv < "$(DDGV)/$(RAW_LAD)" | \
		./normalise -t LAD -c lad17cd -e lad17nm -w lad17nmw \
	'
clean::
	rm -f "$(STANDARD_LAD)"

$(STANDARD_LSOA): $(DDGV)/$(RAW_LSOA) ./normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
		./normalise -t LSOA -c LSOA11CD -e LSOA11NM -w LSOA11NMW < "$(DDGV)/$(RAW_LSOA)" \
	'
clean::
	rm -f "$(STANDARD_LSOA)"


$(STANDARD_MSOA): $(DDGV)/$(RAW_MSOA) $(DDGV)/$(MSOA_NAMES) rename-msoas normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
		./rename-msoas -n "$(DDGV)/$(MSOA_NAMES)" < "$(DDGV)/$(RAW_MSOA)" | \
		./normalise -t MSOA -c MSOA11CD -e MSOA11NM -e MSOA11NMW \
	'
clean::
	rm -f "$(STANDARD_MSOA)"

$(STANDARD_OA): $(DDGV)/$(RAW_OA) normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
	./normalise -t OA -c OA11CD -e OA11CD -w OA11CD < "$(DDGV)/$(RAW_OA)" \
	'
clean::
	rm -f "$(STANDARD_OA)"
