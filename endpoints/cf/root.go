package cf

import ()

type EndpointCfRoot struct {
	Helper *CfHelper `inject:"y"`
}

func (self *EndpointCfRoot) getHelper() *CfHelper {
	return self.Helper
}

// XXX: We should probably authenticate against CloudFoundry!

func (self *EndpointCfRoot) ItemCatalog() *EndpointCatalog {
	child := &EndpointCatalog{}
	child.Parent = self
	return child
}

// XXX: This needs to map service_instances... may need to use Item
func (self *EndpointCfRoot) ItemServiceInstances() *EndpointServiceInstances {
	child := &EndpointServiceInstances{}
	child.Parent = self
	return child
}
