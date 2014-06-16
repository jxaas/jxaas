package endpoints

import (
	"github.com/justinsb/gova/inject"

	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/model"
)

type EndpointBundle struct {
	Parent     *EndpointBundles
	BundleType bundletype.BundleType
}

func (self *EndpointBundle) Item(key string, injector inject.Injector) *EndpointInstance {
	child := &EndpointInstance{}
	child.Parent = self
	child.InstanceId = key

	injector.Inject(&child.Huddle)

	return child
}

func (self *EndpointBundle) HttpGet(huddle *core.Huddle) ([]*model.Instance, error) {
	tenant := self.Parent.Parent.Tenant
	bundleType := self.BundleType

	instances, err := huddle.ListInstances(tenant, bundleType)
	if err != nil {
		return nil, err
	}
	if instances == nil {
		return nil, nil
	}

	models := []*model.Instance{}
	for _, instance := range instances {
		model, err := instance.GetState()
		if err != nil {
			return nil, err
		}

		models = append(models, model)
	}

	return models, nil
}
