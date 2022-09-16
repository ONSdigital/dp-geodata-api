#! /bin/sh
#
# join2.sh -- join assignment comment lines
#
# To have a variable show up in "make help", put a #= comment line before
# the assignment, like this:
#
#	#= variable description
#	VAR=val

sed -e '
/^#=/ {
	N
	s/\n/#=/
}
' "$@"
