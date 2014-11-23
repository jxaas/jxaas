package router

type SimpleRouterRegistry struct {
	mapping map[string]string
}

func NewSimpleRouterRegistry() *SimpleRouterRegistry {
	self := &SimpleRouterRegistry{}
	return self
}

func (self*SimpleRouterRegistry) GetBackendForTenant(tenant string) string {
	v := self.mapping[tenant]
	return v
}

func (self*SimpleRouterRegistry) SetBackendForTenant(tenant string, backend string) error {
	self.mapping[tenant] = backend
	return nil
}
