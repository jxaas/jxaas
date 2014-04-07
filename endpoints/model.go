package endpoints

import "launchpad.net/juju-core/state/api"

type Instance struct {
	Id string

	Units map[string]*Unit
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

func MapToInstance(id string, api *api.ServiceStatus) *Instance {
	instance := &Instance{}
	instance.Id = id
	instance.Units = make(map[string]*Unit)
	for key, unit := range api.Units {
		instance.Units[key] = MapToUnit(key, &unit)
	}
	return instance
}
