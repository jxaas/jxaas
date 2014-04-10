package bundle

import "fmt"

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

func (self *RelationConfig) String() string {
	return fmt.Sprintf("Relation: [%v -> %v]", self.From, self.To)
}
