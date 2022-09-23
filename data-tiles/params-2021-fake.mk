#
# params-2021-fake.mk -- parameters to generate fake data for 2021 geos
#
# To use this file:
#
#	make all PARAMS=params-2021-fake.mk

# use the bundled quads file
DATA_TILE_GRID=DataTileGrid.json

# use content-2022-09-05.json from confluence
CATVERSION=2021

# use geo-2021.mk
GEOVERSION=2021

# use met-fake.mk
METVERSION=fake

# place output in 2021-fake subdirectory
DO=$D/output/2021-fake
