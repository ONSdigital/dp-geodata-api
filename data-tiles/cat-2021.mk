#
# cat-2011.mk -- rules to get 2011 categories.txt
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	CAT_DOWNLOADS	list of files to download for categories
#	CAT_STANDARD	name of categories.txt
#
# Each file named in those variables should have targets in this file.

# For 2011 categories, all we have is categories.txt in this repo.
CAT_DOWNLOADS=$(DDMV)/content-2022-09-05.json

$(CAT_DOWNLOADS):
	@echo "Please download content.json 2022-09-05 from Confluence"
	@echo "and place in $(CAT_DOWNLOADS)"
	false
realclean::
	@echo "Not removing manually downloaded $(CAT_DOWNLOADS)"

CAT_STANDARD=$(DDMV)/categories.txt

$(CAT_STANDARD): $(CAT_DOWNLOADS)
	./atomic.sh "$(CAT_STANDARD)" ./extract-categories < "$(CAT_DOWNLOADS)"
clean::
	rm -f "$(CAT_STANDARD)"
