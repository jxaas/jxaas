package router

type RouterRegistry interface {
	GetBackendForTenant(tenant string) string
	SetBackendForTenant(tenant string, backend string) error
}
