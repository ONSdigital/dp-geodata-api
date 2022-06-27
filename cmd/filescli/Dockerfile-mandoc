FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install ca-certificates && \
    apt-get -y install mandoc && \
    apt-get clean

VOLUME /source
WORKDIR /source
