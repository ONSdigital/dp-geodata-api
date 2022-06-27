#! /bin/sh

if which mandoc
then
    mandoc "$@"
else
    docker run -i -v `pwd`:/source --rm local/mandoc /usr/bin/mandoc "$@"
fi
