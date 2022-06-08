#! /bin/sh

grep code: | awk '
{
    if (length($2) > 10) {
        cat = substr($2, 2, 11)
        if (!match(cat, /0001$/)) {
            print cat
        }
    }
}
'