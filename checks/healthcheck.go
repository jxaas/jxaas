package checks

import (
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type HealthCheck interface {
	Run(client *juju.Client, serviceId string, repair bool) (*model.Health, error)
}
