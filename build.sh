#!/bin/bash -e

WORKDIR=${PWD}/.build/
export GOBIN=${WORKDIR}/bin
export GOPATH=${WORKDIR}/gopath

if [ ! -h ${GOPATH}/src/github.com/jxaas/jxaas ]; then
	mkdir -p ${GOPATH}/src/github.com/jxaas/
	ln -s ../../../../.. ${GOPATH}/src/github.com/jxaas/jxaas
fi

# Get the dependencies
go get -v -d code.google.com/p/go.crypto/...
go get -v -d code.google.com/p/go.net/...

# Get the trusty version of juju-core
if [ ! -d ${GOPATH}/src/launchpad.net/juju-core ]; then
	mkdir -p ${GOPATH}/src/launchpad.net
	echo "Fetching lp:juju-core/1.18"
	pushd ${GOPATH}/src/launchpad.net/
	bzr branch lp:juju-core/1.18 juju-core
	popd
fi

# More dependencies
go get -v -d github.com/jxaas/jxaas
go get -v -d launchpad.net/juju-core

# Get the trusty version of go.crypto
pushd ${GOPATH}/src/code.google.com/p/go.crypto
hg checkout 191
popd


# Make sure it is installed
go install -v github.com/jxaas/jxaas/...

rm -rf ${WORKDIR}/templates/
cp -r templates/ ${WORKDIR}/templates/

# Build archive
pushd ${WORKDIR}
tar czvf jxaas.tar.gz bin/* templates/*
popd
