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


# Get the trusty version of juju-core
# It would sure be nice if godeps handled go dependencies...
if [ ! -d ${GOPATH}/src/launchpad.net/juju-core ]; then
       mkdir -p ${GOPATH}/src/launchpad.net
       echo "Fetching lp:juju-core/1.18"
       pushd ${GOPATH}/src/launchpad.net/
       bzr branch lp:juju-core/1.18 juju-core
       popd
fi

# Get the godeps tool
go get launchpad.net/godeps

# Switch branches (doesn't seem to be a way to do this in godeps?)
if [[ ! -d ${GOPATH}/src/launchpad.net/goose/ ]]; then
  pushd ${GOPATH}/src/launchpad.net/
  bzr branch lp:~justin-fathomdb/goose/keystone_improvements goose
  popd
fi

# Download, but do not install, the latest code (to seed godeps)
go get -d -u -v github.com/jxaas/jxaas

# Install some dependencies (these are otherwise missed?)
#go get -d github.com/mattbaird/elastigo

# Install the correct versions of dependencies
${GOBIN}/godeps -u dependencies.tsv 

# Make sure it is installed
go install -v github.com/jxaas/jxaas/cmd/jxaas-admin
go install -v github.com/jxaas/jxaas/cmd/jxaas-routerd
go install -v github.com/jxaas/jxaas/cmd/jxaasd

rm -rf ${WORKDIR}/templates/
cp -r templates/ ${WORKDIR}/templates/

# Build archive
pushd ${WORKDIR}
tar czvf jxaas.tar.gz bin/* templates/*
popd
