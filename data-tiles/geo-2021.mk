#
# geo-2021.mk -- rules to download and process 2021 geographies
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
# URLs for raw geojson files
#

# The LAD geodata is found by:
# 1. Go to https://geoportal.statistics.gov.uk/datasets/ons::local-authority-districts-december-2021-uk-bgc/explore?location=55.215503%2C-3.316939%2C7.48
# 2. Select Download
# 3. Select GeoJSON
LAD_URL=https://opendata.arcgis.com/api/v3/datasets/0f131e03df73415b824a1c214594eeab_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

MSOA_URL=https://opendata.arcgis.com/api/v3/datasets/38255434eb54456cbd202f54fddfe5c9_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

OA_URL=https://opendata.arcgis.com/api/v3/datasets/5670c14a21224d8187357a095121ca39_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

#
# paths to downloaded geojson files
#

RAW_LAD=$(DDGV)/lad.geojson
RAW_MSOA=$(DDGV)/msoa.geojson
RAW_OA=$(DDGV)/oa.geojson

#
# download geojson files
#

$(RAW_LAD):
	./atomic.sh "$@" curl "$(LAD_URL)"
clean::
	rm -f "$(RAW_LAD).new"
realclean::
	rm -f "$(RAW_LAD)"

$(RAW_MSOA):
	./atomic.sh "$@" curl "$(MSOA_URL)"
clean::
	rm -f "$(RAW_MSOA).new"
realclean::
	rm -f "$(RAW_MSOA)"

$(RAW_OA):
	./atomic.sh "$@" curl "$(OA_URL)"
clean::
	rm -f "$(RAW_OA).new"
realclean::
	rm -f "$(RAW_OA)"

#
# MSOA names file from House of Commons
#
# URL to download MSOA names
MSOA_NAMES_URL=https://houseofcommonslibrary.github.io/msoanames/MSOA-Names-2.0.csv

# path to downloaded msoa names file
MSOA_NAMES=$(DDGV)/msoa-names.csv

$(MSOA_NAMES):
	./atomic.sh "$@" curl "$(MSOA_NAMES_URL)"
clean::
	rm -f "$(MSOA_NAMES).new"
realclean::
	rm -f "$(MSOA_NAMES)"


#
# Set GEO_DOWNLOADS so parent makefile can use download files as targets.
#

# paths to local downloaded raw geojson files and MSOA names file
GEO_DOWNLOADS=$(RAW_LAD) $(RAW_MSOA) $(RAW_OA) $(MSOA_NAMES)


#
# Processed geo files
#

# The files in $DPG are versions of the downloaded geojson, but with bboxes added
# for each feature, and with geotype, geocode, ename and wname properties added.
#
# Also MSOA names are added, and certain LAD names are changed.

STANDARD_LAD=$(DPGV)/lad.geojson

STANDARD_MSOA=$(DPGV)/msoa.geojson

STANDARD_OA=$(DPGV)/oa.geojson

#
# Set GEO_STANDARD so parent make file can use processed files as targets.
#
GEO_STANDARD=$(STANDARD_LAD) $(STANDARD_MSOA) $(STANDARD_OA)

#
# Rules to process geo files
#

$(STANDARD_LAD): $(RAW_LAD) normalise
	./atomic.sh "$@" ./normalise -t LAD -c LAD21CD -e LAD21NM -w LAD21NMW < "$(RAW_LAD)"
clean::
	rm -f "$(STANDARD_LAD).new" "$(STANDARD_LAD)"

$(STANDARD_MSOA): $(RAW_MSOA) $(MSOA_NAMES) rename-msoas normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
		./rename-msoas \
			-c msoa21cd \
			-e msoa21hclnm \
			-w msoa21hclnmw \
			-C MSOA21CD \
			-E MSOA21NM \
			-W MSOA21NMW \
			-n "$(MSOA_NAMES)" < "$(RAW_MSOA)" | \
		./normalise -t MSOA -c MSOA21CD -e MSOA21NM -w MSOA21NM \
	'
clean::
	rm -f "$(STANDARD_MSOA).new" "$(STANDARD_MSOA)"

$(STANDARD_OA): $(RAW_OA) normalise
	./atomic.sh "$@" ./normalise -t OA -c OA21CD -e OA21CD -w OA21CD < "$(RAW_OA)"
clean::
	rm -f "$(STANDARD_OA).new" "$(STANDARD_OA)"
