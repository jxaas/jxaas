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

// XXX: This needs to map service_instances... may need to use Item
func (self *EndpointCfV2) ItemServiceInstances() *EndpointServiceInstances {
	child := &EndpointServiceInstances{}
	child.Parent = self
	return child
}
