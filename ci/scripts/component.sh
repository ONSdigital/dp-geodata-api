#!/bin/bash -eux

pushd dp-geodata-api
  make test-component
popd
