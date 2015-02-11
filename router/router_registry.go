package router

import (
	"github.com/jxaas/jxaas/model"
)

type RouterRegistry interface {
	ListServices() ([]string, error)

	GetBackendForTenant(service string, tenant *string) string
	ListServicesForTenant(tenant string) (*model.Bundles, error)
	SetBackendForTenant(service string, tenant string, backend string) error
	SetBackendForService(service string, backend string) error
}
