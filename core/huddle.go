package core

import (
	"net"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/juju"
)

// System is the top-level object for storing system state
type System struct {
	BundleTypes map[string]bundletype.BundleType
}

// Gets the bundle type by key
func (self *System) GetBundleType(key string) bundletype.BundleType {
	return self.BundleTypes[key]
}

// A Huddle is a group of servers. For us, it is a Juju environment into which multiple tenants are deployed.
// Some services are shared across the huddle.
type Huddle struct {
	// URL for the private API (the stub uses this to call private API functions)
	PrivateUrl string

	System         *System
	SharedServices map[string]*SharedService

	JujuClient *juju.Client
}

// Implement fmt.Stringer
func (self *Huddle) String() string {
	return log.AsJson(self)
}

// A Juju service that is used by multiple JXaaS instances
// Used, for example, for logging/monitoring services.
type SharedService struct {
	Key           string
	JujuName      string
	PublicAddress net.IP
}

// Implement fmt.Stringer
func (self *SharedService) String() string {
	return log.AsJson(self)
}

// Returns the URL base for the private API server
func (self *Huddle) GetPrivateUrl() string {
	return self.PrivateUrl
}
