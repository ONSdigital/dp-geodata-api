#
# met-2011.mk -- rules to download and process 2011 metrics
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	MET_DOWNLOADS	list of downloaded metrics files
#	MET_PROCESSED	list of processed standard format metrics files
#
# Each file named in those variables should have targets in this file.

# Location of nomis zips
NOMIS_URL=https://www.nomisweb.co.uk/output/census/2011

# List of nomis zips to download, taken from Viv's v4 google spreadsheet.
NOMIS_ZIPS=\
	ks103ew_2011_oa.zip \
	ks202ew_2011_oa.zip \
	ks206ew_2011_oa.zip \
	ks207wa_2011_oa.zip \
	ks608ew_2011_oa.zip \
	qs101ew_2011_oa.zip \
	qs103ew_2011_oa.zip \
	qs104ew_2011_oa.zip \
	qs113ew_2011_oa.zip \
	qs119ew_2011_oa.zip \
	qs201ew_2011_oa.zip \
	qs202ew_2011_oa.zip \
	qs203ew_2011_oa.zip \
	qs208ew_2011_oa.zip \
	qs301ew_2011_oa.zip \
	qs302ew_2011_oa.zip \
	qs303ew_2011_oa.zip \
	qs402ew_2011_oa.zip \
	qs403ew_2011_oa.zip \
	qs406ew_2011_oa.zip \
	qs411ew_2011_oa.zip \
	qs415ew_2011_oa.zip \
	qs416ew_2011_oa.zip \
	qs501ew_2011_oa.zip \
	qs601ew_2011_oa.zip \
	qs604ew_2011_oa.zip \
	qs605ew_2011_oa.zip \
	qs701ew_2011_oa.zip \
	qs702ew_oa.zip \
	qs803ew_2011_oa.zip


# How to download .zip files
%.zip:
	./atomic.sh "$@" curl $(NOMIS_URL)/`basename "$@"`

# paths to local downloaded zip files
MET_DOWNLOADS=$(foreach zip,$(NOMIS_ZIPS),$(DDMV)/$(zip))

realclean::
	rm -f $(foreach path,$(MET_DOWNLOADS),"$(path)")


# Zip files are extracted into a directory with an .extracted suffix.
MET_PROCESSED=$(patsubst %.zip,%.extracted,$(MET_DOWNLOADS))

# how to extract a zip file
$(MET_PROCESSED): %.extracted: %.zip
	./atomic-rm.sh "$@".tmp
	unzip -d "$@".tmp "$<"
	mv "$@".tmp "$@"

clean::
	./atomic-rm.sh $(foreach path,$(MET_PROCESSED),"$(path)")
