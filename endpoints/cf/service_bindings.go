package cf

import ()

type EndpointServiceBindings struct {
	Parent *EndpointServiceInstance
}

func (self *EndpointServiceBindings) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointServiceBindings) getInstanceId() string {
	return self.Parent.getInstanceId()
}

func (self *EndpointServiceBindings) Item(key string) *EndpointServiceBinding {
	child := &EndpointServiceBinding{}
	child.Parent = self
	child.Id = key
	return child
}
