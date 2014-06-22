package endpoints

import (
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/model"
)

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

func (self *EndpointBundles) HttpGet(huddle *core.Huddle) (*model.Bundles, error) {
	bundles := huddle.System.ListBundleTypes()
	return bundles, nil
}
