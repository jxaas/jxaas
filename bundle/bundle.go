package bundle

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
