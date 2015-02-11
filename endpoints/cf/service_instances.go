package cf

type EndpointServiceInstances struct {
	Parent *EndpointCfV2
}

func (self *EndpointServiceInstances) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointServiceInstances) getService() *EndpointCfService {
	return self.Parent.getService()
}

func (self *EndpointServiceInstances) Item(key string) *EndpointServiceInstance {
	child := &EndpointServiceInstance{}
	child.Parent = self
	child.Id = key
	return child
}
