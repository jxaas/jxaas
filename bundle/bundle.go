package bundle

import (
	"fmt"

	"github.com/justinsb/gova/log"
)

type Bundle struct {
	Services  map[string]*ServiceConfig
	Relations []*RelationConfig
}

type ServiceConfig struct {
	Charm       string
	Branch      string
	NumberUnits int
	Options     map[string]string
	Exposed     bool
}

type RelationConfig struct {
	From string
	To   string
}

// Implement fmt.Stringer
func (self *Bundle) String() string {
	return log.AsJson(self)
}

// Implement fmt.Stringer
func (self *RelationConfig) String() string {
	return fmt.Sprintf("Relation: [%v -> %v]", self.From, self.To)
}
