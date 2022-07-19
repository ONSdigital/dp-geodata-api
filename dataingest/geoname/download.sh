#! /bin/sh

# not sure where these files are used:

for f in ChangeHistory.csv Equivalents.csv
do
    aws --profile dp-sandbox s3 cp s3://ons-dp-sandbox-atlas-input/geoname/$f .
done
curl -o MSOA-Names-1.16.csv https://houseofcommonslibrary.github.io/msoanames/MSOA-Names-1.16.csv
