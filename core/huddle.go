package core

import (
	"encoding/json"
	"fmt"
	"net"
)

// A Huddle is a group of servers. For us, it is a Juju environment into which multiple tenants are deployed.
// Some services are shared across the huddle.
type Huddle struct {
	SharedServices map[string]*SharedService
}

func (self *Huddle) String() string {
	return asJson(self)
}

type SharedService struct {
	Key           string
	JujuName      string
	PublicAddress net.IP
}

func asJson(o interface{}) string {
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
	return asJson(self)
}
