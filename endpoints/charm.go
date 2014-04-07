package endpoints

type EndpointCharm struct {
	Parent      *EndpointServices
	ServiceType string
}

func (self *EndpointCharm) Item(key string) *EndpointService {
	child := &EndpointService{}
	child.Parent = self
	child.ServiceId = key
	return child
}
