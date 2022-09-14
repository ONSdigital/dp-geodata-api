#
# cat-2011.mk -- rules to get 2011 categories.txt
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	CAT_DOWNLOAD	list of files to download for categories
#	CAT_STANDARD	name of categories.txt
#
# Each file named in those variables should have targets in this file.

# For 2011 categories, all we have is categories.txt in this repo.
CAT_DOWNLOADS=categories.txt
CAT_STANDARD=$(CAT_DOWNLOADS)
