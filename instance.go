package jxaas

import (
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

// A JXaaS instance
type Instance interface {
	RunHealthCheck(repair bool) (*model.Health, error)

	GetJujuClient() *juju.Client
}
