#!/bin/bash -eux

pushd dp-geodata-api
  make test
  make check-generate
popd
