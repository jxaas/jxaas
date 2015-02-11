package jxaas

import (
	"github.com/juju/juju/state/api"
	"github.com/jxaas/jxaas/model"
)

type HealthCheck interface {
	Run(instance Instance, jujuState map[string]api.ServiceStatus, repair bool) (*model.Health, error)
}
