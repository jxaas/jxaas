package endpoints

type EndpointTenant struct {
	Tenant string
}

func (self *EndpointTenant) ItemServices() *EndpointBundles {
	child := &EndpointBundles{}
	child.Parent = self
	return child
}
