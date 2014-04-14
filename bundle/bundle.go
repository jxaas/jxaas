package bundle

import (
	"fmt"

	"github.com/jxaas/jxaas/core"
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

func (self *Bundle) String() string {
	return core.AsJson(self)
}

func (self *RelationConfig) String() string {
	return fmt.Sprintf("Relation: [%v -> %v]", self.From, self.To)
}
