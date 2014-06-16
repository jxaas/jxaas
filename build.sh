#!/bin/bash

set -e
set -x

WORKDIR=${PWD}/.build/
export GOBIN=${WORKDIR}/bin
export GOPATH=${WORKDIR}/gopath

if [ ! -h ${GOPATH}/src/github.com/jxaas/jxaas ]; then
	mkdir -p ${GOPATH}/src/github.com/jxaas/
	ln -s ../../../../.. ${GOPATH}/src/github.com/jxaas/jxaas
fi

# Get the godeps tool
go get launchpad.net/godeps

# Download, but do not install, the latest code (to seed godeps)
go get -d -u -v github.com/jxaas/jxaas

# Install the correct versions of dependencies
${GOBIN}/godeps -u dependencies.tsv 

# Make sure it is installed
go install -v github.com/jxaas/jxaas/cmd/jxaasd

rm -rf ${WORKDIR}/templates/
cp -r templates/ ${WORKDIR}/templates/

# Build archive
pushd ${WORKDIR}
tar czvf jxaas.tar.gz bin/* templates/*
popd
