#! /bin/sh

cd out

aws s3 \
    --profile development \
    --region eu-west-2 \
    sync . s3://find-insights-db-dumps/quads
