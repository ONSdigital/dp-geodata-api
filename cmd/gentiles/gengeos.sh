#! /bin/sh

psql -c 'select code from geo where type_id in (4,5,6,7) '|
sed \
    -e '/code/d' \
    -e '/^-/d' \
    -e 's/^ //' \
    -e '/rows/d' \
    -e '/^$/d' |
sort
