package endpoints

type EndpointBundles struct {
	Parent *EndpointTenant
}

func (self *EndpointBundles) Item(key string) *EndpointBundle {
	child := &EndpointBundle{}
	child.Parent = self
	child.BundleType = key

	return child
}
