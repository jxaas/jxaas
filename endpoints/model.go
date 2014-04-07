package endpoints

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/api/params"
)

type Instance struct {
	Id string

	// (Optional), so *bool
	Exposed *bool

	Units map[string]*Unit

	Config map[string]string

	ConfigParameters map[string]ConfigParameter
}

//func (self *Instance) ConfigValues() map[string]string {
//	flat := make(map[string]string)
//	for k, v := range self.Config {
//		flat[k] = v.Value
//	}
//	return flat
//}

type ConfigParameter struct {
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

func MapToConfigParameters(config *params.ServiceGetResults) map[string]ConfigParameter {
	out := make(map[string]ConfigParameter)

	if config.Config != nil {
		for k, v := range config.Config {
			m, ok := v.(map[string]interface{})
			if !ok {
				log.Warn("Unexpected type for config value: %v", k)
				continue
			}

			p := &ConfigParameter{}
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

func MapToConfig(config *params.ServiceGetResults) map[string]string {
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

func MapToInstance(id string, api *api.ServiceStatus, config *params.ServiceGetResults) *Instance {
	instance := &Instance{}
	instance.Id = id
	instance.Units = make(map[string]*Unit)
	instance.Exposed = &api.Exposed

	for key, unit := range api.Units {
		instance.Units[key] = MapToUnit(key, &unit)
	}
	if config != nil {
		instance.Config = MapToConfig(config)
		instance.ConfigParameters = MapToConfigParameters(config)
	}

	return instance
}
