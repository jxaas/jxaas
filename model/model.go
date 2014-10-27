package model

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/api/params"
)

type RelationInfo struct {
	Properties      map[string]string
	Timestamp       string
	PublicAddresses []string
}

type Instance struct {
	Id string

	// (Optional), so *bool
	Exposed *bool

	// (Optional)
	NumberUnits *int

	Status string

	Units map[string]*Unit

	// The configuration options, as set by the user.
	// Some of these values will map directly to Juju instance config,
	// some will be jxaas-specified or transformed.
	Options map[string]string

	OptionDescriptions map[string]OptionDescription
}

type OptionDescription struct {
	Default     string
	Description string
	Type        string
}

type Unit struct {
	Id string

	PublicAddress string

	Status string
}

func MapToUnit(id string, api *api.UnitStatus) *Unit {
	unit := &Unit{}
	unit.Id = id
	unit.PublicAddress = api.PublicAddress
	unit.Status = string(api.AgentState)
	return unit
}

func getString(m map[string]interface{}, key string) string {
	v, found := m[key]
	if !found {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		s := fmt.Sprint(v)

		//	log.Warn("Expected string for key: %v, was %v", key, reflect.TypeOf(v))
		return s
	}

	return s
}

func mapToOptionDescriptions(config *params.ServiceGetResults) map[string]OptionDescription {
	out := make(map[string]OptionDescription)

	if config.Config != nil {
		for k, v := range config.Config {
			m, ok := v.(map[string]interface{})
			if !ok {
				log.Warn("Unexpected type for config value: %v", k)
				continue
			}

			p := &OptionDescription{}
			p.Type = getString(m, "type")
			p.Description = getString(m, "description")

			// juju returns true if the value is the default, false otherwise,
			// but does not return the actual default value.  That's uninintuitive to me,
			// so block it.
			//p.Default = getString(m, "default")

			out[k] = *p
		}
	}

	return out
}

func mapToOptions(config *params.ServiceGetResults) map[string]string {
	out := make(map[string]string)

	if config.Config != nil {
		for k, v := range config.Config {
			m, ok := v.(map[string]interface{})
			if !ok {
				log.Warn("Unexpected type for config value: %v", k)
				continue
			}

			out[k] = getString(m, "value")
		}
	}

	return out
}

func MergeInstanceStatus(instance *Instance, unit *api.UnitStatus) {
	unitStatus := string(unit.AgentState)

	if instance.Status != unitStatus {
		if instance.Status == "" {
			instance.Status = unitStatus
		} else {
			// TODO: Resolve mixed state
			log.Warn("Unable to resolve mixed state: %v vs %v", instance.Status, unitStatus)
		}
	}
}

func MapToInstance(id string, api *api.ServiceStatus, config *params.ServiceGetResults) *Instance {
	instance := &Instance{}
	instance.Id = id

	if api != nil {
		instance.Exposed = &api.Exposed

		instance.Units = make(map[string]*Unit)
		for key, unit := range api.Units {
			unitState := MapToUnit(key, &unit)
			instance.Units[key] = unitState

			MergeInstanceStatus(instance, &unit)
		}

		instance.NumberUnits = new(int)
		*instance.NumberUnits = len(api.Units)
	}

	if config != nil {
		instance.Options = mapToOptions(config)
		instance.OptionDescriptions = mapToOptionDescriptions(config)
	}

	return instance
}

type RelationProperty struct {
	UnitId string

	RelationType string
	RelationKey  string

	Key   string
	Value string
}

func (self *RelationProperty) String() string {
	return log.AsJson(self)
}
