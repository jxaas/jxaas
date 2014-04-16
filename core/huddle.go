package core

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jxaas/jxaas/juju"
)

// A Huddle is a group of servers. For us, it is a Juju environment into which multiple tenants are deployed.
// Some services are shared across the huddle.
type Huddle struct {
	SharedServices map[string]*SharedService

	JujuClient *juju.Client
}

func (self *Huddle) String() string {
	return AsJson(self)
}

type SharedService struct {
	Key           string
	JujuName      string
	PublicAddress net.IP
}

func AsJson(o interface{}) string {
	if o == nil {
		return "nil"
	}
	bytes, err := json.Marshal(o)
	if err != nil {
		return fmt.Sprintf("[Error: %v]", err)
	}

	return string(bytes)
}

func (self *SharedService) String() string {
	return AsJson(self)
}
