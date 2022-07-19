#! /bin/sh

cd out

aws s3 \
    --profile dp-sandbox \
    sync . s3://ons-dp-sandbox-atlas-data/quads
