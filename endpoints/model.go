package endpoints

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/api/params"
)

type Instance struct {
	Id string

	Units map[string]*Unit

	Config map[string]ConfigValue
}

//func (self *Instance) ConfigValues() map[string]string {
//	flat := make(map[string]string)
//	for k, v := range self.Config {
//		flat[k] = v.Value
//	}
//	return flat
//}

type ConfigValue struct {
	Default     string
	Description string
	Type        string
	Value       string
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

func MapToConfiguration(config *params.ServiceGetResults) map[string]ConfigValue {
	out := make(map[string]ConfigValue)

	if config.Config != nil {
		for k, v := range config.Config {
			m, ok := v.(map[string]interface{})
			if !ok {
				log.Warn("Unexpected type for config value: %v", k)
				continue
			}

			configValue := &ConfigValue{}
			configValue.Type = getString(m, "type")
			configValue.Description = getString(m, "description")
			configValue.Default = getString(m, "default")
			configValue.Value = getString(m, "value")

			out[k] = *configValue
		}
	}

	return out
}

func MapToInstance(id string, api *api.ServiceStatus, config *params.ServiceGetResults) *Instance {
	instance := &Instance{}
	instance.Id = id
	instance.Units = make(map[string]*Unit)
	for key, unit := range api.Units {
		instance.Units[key] = MapToUnit(key, &unit)
	}
	if config != nil {
		instance.Config = MapToConfiguration(config)
	}

	return instance
}
