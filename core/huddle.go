package core

import (
	"net"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/juju"
)

type System struct {
	BundleStore *bundle.BundleStore
}

// A Huddle is a group of servers. For us, it is a Juju environment into which multiple tenants are deployed.
// Some services are shared across the huddle.
type Huddle struct {
	System         *System
	SharedServices map[string]*SharedService

	JujuClient *juju.Client
}

func (self *Huddle) String() string {
	return log.AsJson(self)
}

type SharedService struct {
	Key           string
	JujuName      string
	PublicAddress net.IP
}

func (self *SharedService) String() string {
	return log.AsJson(self)
}
