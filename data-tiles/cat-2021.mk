#
# cat-2021.mk -- rules to get 2021 categories.txt
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	CAT_DOWNLOADS	list of files to download for categories
#	CAT_STANDARD	name of categories.txt
#
# Each file named in those variables should have targets in this file.

content-2022-09-05.json:
	@echo "Please download content.json 2022-09-05 from Confluence"
	@echo "and place in $?"
	false

CAT_DOWNLOADS=$(DDMV)/content.json
$(CAT_DOWNLOADS): content-2022-09-05.json
	cp content-2022-09-05.json "$(CAT_DOWNLOADS)"
clean::
	rm -f "$(CAT_DOWNLOADS)"

CAT_STANDARD=$(DDMV)/categories.txt

$(CAT_STANDARD): $(CAT_DOWNLOADS) extract-categories
	./atomic.sh "$(CAT_STANDARD)" ./extract-categories < "$(CAT_DOWNLOADS)"
clean::
	rm -f "$(CAT_STANDARD)"
