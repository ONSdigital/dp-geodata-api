#
# met-fake.mk -- rules to generate fake metrics
#
# This makefile is mean to be included by GNUmakefile; it's not independent
#
# Variables this file is expected to set
#
#	MET_DOWNLOADS	list of downloaded metrics files
#	MET_STANDARD	list of normalised standard format metrics files
#
# Each file named in those variables should have targets in this file.
# And there should be clean and realclean targets to clean these files up.

MET_DOWNLOADS=$(DDMV)/DATA.CSV

$(MET_DOWNLOADS): fake-data $(GEO_STANDARD) $(CAT_STANDARD)
	./atomic.sh "$(MET_DOWNLOADS)" ./fake-data -c "$(CAT_STANDARD)" -G "$(DPGV)"

realclean::
	rm -f "$(MET_DOWNLOADS)"

MET_STANDARD=$(MET_DOWNLOADS)
