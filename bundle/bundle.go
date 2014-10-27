package bundle

import (
	"fmt"

	"github.com/justinsb/gova/log"
)

type Bundle struct {
	Services     map[string]*ServiceConfig
	Relations    []*RelationConfig
	Provides     map[string]*ProvideConfig
	HealthChecks map[string]*HealthCheckConfig

	CloudFoundryConfig *CloudFoundryConfig
}

type ServiceConfig struct {
	Charm       string
	Branch      string
	NumberUnits int
	Options     map[string]string
	Exposed     bool
	//OpenPorts   []string
}

type RelationConfig struct {
	From string
	To   string
}

type ProvideConfig struct {
	Properties map[string]string
}

type HealthCheckConfig struct {
	Service string
}

type CloudFoundryConfig struct {
	Credentials map[string]string
}

// Implement fmt.Stringer
func (self *Bundle) String() string {
	return log.AsJson(self)
}

// Implement fmt.Stringer
func (self *RelationConfig) String() string {
	return fmt.Sprintf("Relation: [%v -> %v]", self.From, self.To)
}
