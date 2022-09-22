#
# params-2011.mk -- parameters to generate 2011 data tiles/breaks/geos
#
# To use this file:
#
#	make all PARAMS=params-2011.mk

# use the bundled quads file
DATA_TILE_GRID=DataTileGrid.json

# use cat-2011.mk
CATVERSION=2011

# use geo-2011.mk
GEOVERSION=2011

# use met-2011.mk
METVERSION=2011

# place output in 2011 subdirectory
DO=$D/output/2011
