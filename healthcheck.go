package jxaas

import "github.com/jxaas/jxaas/model"

type HealthCheck interface {
	Run(instance Instance, repair bool) (*model.Health, error)
}
