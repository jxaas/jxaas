package jxaas

import (
	"github.com/jxaas/jxaas/model"
	"github.com/juju/juju/state/api"
)

type HealthCheck interface {
	Run(instance Instance, jujuState map[string]api.ServiceStatus, repair bool) (*model.Health, error)
}
