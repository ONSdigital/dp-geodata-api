#!/bin/bash -eux

pushd dp-geodata-api
  make build
  cp build/dp-geodata-api Dockerfile.concourse ../build
popd
