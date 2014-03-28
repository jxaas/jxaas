
# TODO: Install latest golang

sudo apt-get install mongodb

sudo apt-get install rsyslog-gnutls

# On trusty: (installs a bundled mongodb - ?why?)

sudo apt-get install juju-mongodb

sudo apt-get install lxc bridge-utils

brctl addbr lxcbr0

ip a a 100.64.64.0/24 brd + dev lxcbr0

mkdir ~/juju
cd ~/juju

BASE=`pwd`

export PATH=${BASE}/bin:$PATH
export PATH=$PATH:/usr/local/bin
export GOPATH=${BASE}
export GOHOME=/usr/local/go


go get -v launchpad.net/juju-core
go get -u -v launchpad.net/juju-core/...

go install launchpad.net/juju-core/cmd/...

# Use a sandbox
export JUJU_HOME=${BASE}/.juju/

# Let's use local (LXC)
juju generate-config
juju switch local

sudo JUJU_HOME=${BASE}/.juju ${BASE}/bin/juju bootstrap
