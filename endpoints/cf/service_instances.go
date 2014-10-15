package cf

import ()

type EndpointServiceInstances struct {
	Parent *EndpointCfRoot
}

func (self *EndpointServiceInstances) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointServiceInstances) Item(key string) *EndpointServiceInstance {
	child := &EndpointServiceInstance{}
	child.Parent = self
	child.Id = key
	return child
}
