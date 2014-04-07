package endpoints

type EndpointXaas struct {
}

func (self *EndpointXaas) Item(key string) *EndpointTenant {
	child := &EndpointTenant{}
	child.Tenant = key
	return child
}
