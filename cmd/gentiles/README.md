# Generate Data Tiles

## Generate Category Listing

1. Get the current `src/data/content.ts` file from the `dp-census-atlas` repo.

2. Run

        ./gencats.sh < content.ts > categories.txt

## Get Tile Bounding Boxes

1. Get the current `data-tile-grids/quadsDataTileGrids.json` file from the `dp-census-atlas` repo.

## Generate Geocode Listing

1. Ensure your `PG*` environment variables are set up to talk to the right postgres instance.

2. Run

        ./gengeos.sh > geos.txt

## Generate Data Tile Files

_This takes a long time and can make your CPU very hot._

1. Ensure your `PG*` environment variables are set up to talk to the right postgres intance.

2. Run

        go run ./main.go -C categories.txt -T quadsDataTileGrids.json -G geos.txt -j 10 -o out 2>main.stderr

    The `-j` option is concurrency.
    Low concurrency takes a long time, but high concurrency runs very hot.

    You can `^C` and restart any time, so you can try different concurrencies, or give your machine a rest to cool down.

    The redirection to `main.stderr` removes noise from your stdout so you can just see progress reports.

## Upload Content to S3

1. Ensure you can run the AWS CLI to reach resources in the develop environment.

2. Verify `upload.sh` names the correct S3 bucket and prefix, and add `--delete` and/or `--dryrun` as needed.

3. Run

        ./upload.sh

    This can take a while.

4. Fix permissions on the new files in the UI so they are readable by public.
