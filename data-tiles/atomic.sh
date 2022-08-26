#! /bin/sh
#
# atomic.sh -- capture process stdout to file iff process exits 0

usage() {
	echo "usage: $0 out cmd [arg ...]" 1>&2
	exit $1
}

main() {
	local out=$1
	local tmp
	shift

	if test -z "$out" -o -z "$1"
	then
		usage 1
	fi

	tmp="$out".new
	exec 1> "$tmp" || exit
	"$@"
	status=$?
	if test $status != 0
	then
		exit $status
	fi
	mv "$tmp" "$out"
}

main "$@"
