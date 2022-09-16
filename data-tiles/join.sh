#! /bin/sh
#
# join.sh -- join backslash-continued lines

sed -e '
:x
/\\$/ {
	N
	s/\\\n//g
	bx
}' "$@"
