This directory holds files related to the non-postgres, cli-based data tile generator.

Prerequisites:

	bash
	GNU make
	Go compiler
	unzip
	curl
	aws cli and SSO login details

To build output files from scratch:

	export AWS_PROFILE=dp-sandbox
	aws sso login --profile dp-sandbox

	make			# downloads, processes, generates output files

The AWS login is only needed when downloading. Once the raw files have been downloaded,
you don't need the aws cli or the environment variable.

The output files will be in data/output/

Overall flow of the process:

	Download (make download)
		Files are downloaded and placed in data/downloads/{geo,metrics}

	Extract (make extract)
		Zip files holding metrics are extracted into
		data/downloads/metrics/*.extracted directories.

	Normalise (make standard)
		The downloaded geojson files are normalised and placed in
		data/processed/geo.
		Bounding box features are added, and certain standard property names
		are set: geotype, geocode, ename, wname.
		Also the House of Commons MSOA names are added to the msoa geojson.

	Split metrics (make metrics)
		The extracted zip files are split into files holding single categories.
		The results are placed in data/processed/metrics.

	Generate geometry files (make geos)
		The normalised geojson files are split into individual geojson files
		for each feature.
		Results are placed in data/output/geo.

	Generate data tiles (make tiles)
		content.json and categories are used to select data from the split
		metrics and normalised geojson files.
		Results are placed in data/output/tiles.

	Generate breaks (make breaks)
		ckmeans break files are generated and placed in data/output/breaks.

You don't always have to use individual targets. Most of the time you can just make.
Operations are atomic and dependencies are explicit.

Cleanup targets

	make clean	-- remove binaries and all generated data files

	make realclean	-- same as make clean, but also removes downloaded files

Configuration Files

	categories.txt		-- list of categories to use in output files
	content.json		-- list of quads for generating data tiles
	recode-lads.csv		-- adjustments to LAD names

Data directories

	The make dirs target creates the data directories that are expected to
	exist.
	You should only have to do this once.

Binaries

	Output files depend on their input files, and on the binaries that produce
	them.

	We cannot fully automate rebuilding the binaries when their sources change
	because the Go ecosystem maintains its own dependency tracking, and this
	doesn't integrate well with Make.

	So whenever you change the Go sources, you have to force-rebuild the binary
	with -B.

	You can force-rebuild all binaries:

		make -B binaries

	Or you can recompile individual binaries with the following targets using -B:

		generate-breaks
		generate-tiles
		normalise
		recode-lads
		rename-msoas
		split-geojson
		split-metrics

Atomic operations

	Some steps in the processing take a long time or deal with directories full of
	files.
	We don't want files or directories to be left in an incomplete state, and we
	want Make to be able to pickup where it left off.
	All file creating and directory filling operations must be atomic.

	atomic.sh
		This script takes stdout from another process and writes it atomically
		to the named output file.
		If the process is killed or if it does not exit 0, then the named
		output file is not touched.
		The output file will only be updated if the process exits 0.

	atomic-rm.sh
		This script atomically removes a directory.
		If a directory remove operation is stopped prematurely, the directory
		will be left in an incomplete state and Make would not know about it.
		So this script renames the directory with a .condemned suffix and then
		removes the renamed directory.
		The rename is atomic on a posix filesystem.
		It is always ok to remove a .condemned directory whenever you see one.
