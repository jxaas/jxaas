package endpoints

import (
	"fmt"

	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/api/params"
)

type Instance struct {
	Id string

	Units map[string]*Unit

	Config map[string]string
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

func MapToConfiguration(config map[string]interface{}) map[string]string {
	out := make(map[string]string)

	for k, v := range config {
		out[k] = fmt.Sprint(v)
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
		instance.Config = MapToConfiguration(config.Config)
	}

	return instance
}
