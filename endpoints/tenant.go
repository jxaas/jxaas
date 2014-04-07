package endpoints

type EndpointTenant struct {
	Tenant string
}

func (self *EndpointTenant) ItemServices() *EndpointServices {
	child := &EndpointServices{}
	child.Parent = self
	return child
}
