package endpoints

import "github.com/jxaas/jxaas/core"

type EndpointBundles struct {
	Parent *EndpointTenant
}

func (self *EndpointBundles) Item(key string, huddle *core.Huddle) *EndpointBundle {
	child := &EndpointBundle{}
	child.Parent = self

	bundleType := huddle.System.GetBundleType(key)
	if bundleType == nil {
		return nil
	}

	child.BundleType = bundleType

	return child
}
