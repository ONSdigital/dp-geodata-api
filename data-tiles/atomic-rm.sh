#! /bin/sh
#
# atomic-rm -- remove each directory in a single step to avoid partial contents

usage() {
	echo "usage: $0 <dir> ..." 1>&2
	exit $1
}

remove() {
	local d=$1

	if test -z "$d"
	then
		echo "$0: empty directory name" 1>&2
		exit 1
	fi
	rm -rf "$d".condemned	# remove incomplete from last time
	if ! test -e "$d"
	then
		return 0
	fi
	if ! test -d "$d"
	then
		echo "$0: $d is not a directory; not removing" 1>&2
		exit 1
	fi
	mv "$d" "$d".condemned
	rm -rf "$d".condemned
}

main() {
	local d

	if test -z "$1"
	then
		usage 2
	fi

	for d
	do
		remove "$d"
	done
}

main "$@"
