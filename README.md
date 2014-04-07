
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

JUJU_HOME=${BASE}/.juju ${BASE}/bin/juju bootstrap

# This starts a process:
# /home/justinsb/juju/.juju/local/tools/machine-0/jujud machine --data-dir /home/justinsb/juju/.juju/local --machine-id 0 --debug
# /usr/lib/juju/bin/mongod --auth --dbpath=/home/justinsb/juju/.juju/local/db --sslOnNormalPorts --sslPEMKeyFile /home/justinsb/juju/.juju/local/server.pem --sslPEMKeyPassword xxxxxxx --bind_ip 0.0.0.0 --port 37017 --noprealloc --syslog --smallfiles
# We might need to restart this manually on reboot?
