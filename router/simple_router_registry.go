package router

type SimpleRouterRegistry struct {
	services map[string]string
	tenants map[string]map[string]string
}

func NewSimpleRouterRegistry() *SimpleRouterRegistry {
	self := &SimpleRouterRegistry{}
	return self
}

func (self*SimpleRouterRegistry) GetBackendForTenant(service string, tenant string) string {
	v := self.tenants[tenant][service]
	if v == "" {
		v = self.services[service]
	}
	return v
}

func (self*SimpleRouterRegistry) SetBackendForTenant(service string, tenant string, backend string) error {
	servicesForTenant := self.tenants[tenant]
	if servicesForTenant == nil {
		servicesForTenant = map[string]string{}
		self.tenants[tenant] = servicesForTenant
	}
	servicesForTenant[service] = backend
	return nil
}

func (self*SimpleRouterRegistry) SetBackendForService(service string, backend string) error {
	self.services[service] = backend
	return nil
}

func (self*SimpleRouterRegistry) ListServiceBackends() map[string]string {
	return self.services
}


