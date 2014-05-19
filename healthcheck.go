package jxaas

import (
	"github.com/jxaas/jxaas/model"
	"launchpad.net/juju-core/state/api"
)

type HealthCheck interface {
	Run(instance Instance, jujuState map[string]api.ServiceStatus, repair bool) (*model.Health, error)
}
