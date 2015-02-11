package cf

import ()

type EndpointCfV2 struct {
	Parent *EndpointCfService
}

func (self *EndpointCfV2) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointCfV2) getService() *EndpointCfService {
	return self.Parent.getService()
}

func (self *EndpointCfV2) ItemCatalog() *EndpointCatalog {
	child := &EndpointCatalog{}
	child.Parent = self
	return child
}

func (self *EndpointCfV2) ItemServiceInstances() *EndpointServiceInstances {
	child := &EndpointServiceInstances{}
	child.Parent = self
	return child
}
