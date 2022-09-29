#
# cat-2021-all.mk -- rules to get 2021 content.json
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	CAT_DOWNLOADS	list of files to download for categories
#	CAT_STANDARD	name of categories.txt
#
# Each file named in those variables should have targets in this file.

CAT_2021_ALL_URL=https://ons-dp-sandbox-atlas-data.s3.eu-west-2.amazonaws.com/content-json/2021-all-atlas-content-incomplete-short-var-descs.json

CAT_DOWNLOADS=$(DDCV)/content.json
$(CAT_DOWNLOADS):
	./atomic.sh "$(CAT_DOWNLOADS)" curl "$(CAT_2021_ALL_URL)"
clean::
	rm -f "$(CAT_DOWNLOADS)"

CAT_STANDARD=$(DPCV)/categories.txt

$(CAT_STANDARD): $(CAT_DOWNLOADS) extract-categories
	./atomic.sh "$(CAT_STANDARD)" ./extract-categories < "$(CAT_DOWNLOADS)"
clean::
	rm -f "$(CAT_STANDARD)"
