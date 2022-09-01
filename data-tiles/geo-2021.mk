#
# geo-2021.mk -- rules to download and process 2021 geographies
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	GEO_DOWNLOADS	list of downloaded geo files
#	GEO_PROCESSED	list of processed geo files
#
# Each file named in those variables should have targets in this file.

#
# URLs for raw geojson files
#

LAD_URL=https://opendata.arcgis.com/api/v3/datasets/d31f826e318d441390f54f472d976ee1_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

LSOA_URL=https://opendata.arcgis.com/api/v3/datasets/8fb6ac064a704cee9a999ee08414d61e_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

MSOA_URL=https://opendata.arcgis.com/api/v3/datasets/ec053fbfee484bdcacc3ff8f328f605e_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

OA_URL=https://opendata.arcgis.com/api/v3/datasets/5670c14a21224d8187357a095121ca39_0/downloads/data?format=geojson&spatialRefId=4326&where=1%3D1

#
# paths to downloaded geojson files
#

RAW_LAD=$(DDGV)/lad.geojson
RAW_LSOA=$(DDGV)/lsoa.geojson
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

$(RAW_LSOA):
	./atomic.sh "$@" curl "$(LSOA_URL)"
clean::
	rm -f "$(RAW_LSOA).new"
realclean::
	rm -f "$(RAW_LSOA)"

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
MSOA_NAMES_URL=https://houseofcommonslibrary.github.io/msoanames/MSOA-Names-1.16.csv

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
GEO_DOWNLOADS=$(RAW_LAD) $(RAW_LSOA) $(RAW_MSOA) $(RAW_OA) $(MSOA_NAMES)


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
# Set GEO_PROCESSED so parent make file can use processed files as targets.
#
GEO_PROCESSED=$(STANDARD_LAD) $(STANDARD_LSOA) $(STANDARD_MSOA) $(STANDARD_OA)

#
# Rules to process geo files
#

$(STANDARD_LAD): $(RAW_LAD) normalise
	./atomic.sh "$@" ./normalise -t LAD -c LAD21CD -e LAD21NM -w LAD21NMW < "$(RAW_LAD)"
clean::
	rm -f "$(STANDARD_LAD).new" "$(STANDARD_LAD)"

$(STANDARD_LSOA): $(RAW_LSOA) normalise
	./atomic.sh "$@" ./normalise -t LSOA -c LSOA21CD -e LSOA21NM -w LSOA21NM < "$(RAW_LSOA)"
clean::
	rm -f "$(STANDARD_LSOA).new" "$(STANDARD_LSOA)"


$(STANDARD_MSOA): $(RAW_MSOA) $(MSOA_NAMES) rename-msoas normalise
	./atomic.sh "$@" bash -o pipefail -c ' \
		./rename-msoas -n "$(MSOA_NAMES)" < "$(RAW_MSOA)" | \
		./normalise -t MSOA -c MSOA21CD -e MSOA21NM -w MSOA21NM \
	'
clean::
	rm -f "$(STANDARD_MSOA).new" "$(STANDARD_MSOA)"

$(STANDARD_OA): $(RAW_OA) normalise
	./atomic.sh "$@" ./normalise -t OA -c OA21CD -e OA21CD -w OA21CD < "$(RAW_OA)"
clean::
	rm -f "$(STANDARD_OA).new" "$(STANDARD_OA)"