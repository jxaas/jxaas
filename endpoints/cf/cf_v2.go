package cf

import ()

type EndpointCfV2 struct {
	Parent *EndpointCfRoot
}

func (self *EndpointCfV2) getHelper() *CfHelper {
	return self.Parent.getHelper()
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
