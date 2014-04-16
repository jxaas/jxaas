package endpoints

type EndpointServices struct {
	Parent *EndpointTenant
}

func (self *EndpointServices) Item(key string) *EndpointCharm {
	child := &EndpointCharm{}
	child.Parent = self
	child.ServiceType = key

	return child
}
