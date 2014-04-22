package endpoints

import (
	"net/http"
	"strings"

	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/inject"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"
)

type EndpointBundle struct {
	Parent     *EndpointBundles
	BundleType bundletype.BundleType
}

func (self *EndpointBundle) jujuPrefix() string {
	tenant := self.Parent.Parent.Tenant
	tenant = strings.Replace(tenant, "-", "", -1)

	bundleType := self.BundleType

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + bundleType.Key() + "-"

	return prefix
}

func (self *EndpointBundle) Item(key string, injector inject.Injector) *EndpointInstance {
	child := &EndpointInstance{}
	child.Parent = self
	child.InstanceId = key

	injector.Inject(&child.Huddle)

	return child
}

func (self *EndpointBundle) HttpGet(apiclient *juju.Client) ([]*model.Instance, error) {
	prefix := self.jujuPrefix()

	statuses, err := apiclient.GetServiceStatusList(prefix)
	if err != nil {
		return nil, err
	}
	if statuses == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	instances := make([]*model.Instance, 0)
	for key, state := range statuses {
		// TODO: Make this better - actively match
		// TODO: Reverse the config & shared logic with service get
		if !strings.HasSuffix(key, "-"+self.BundleType.PrimaryJujuService()) {
			continue
		}

		//fmt.Printf("%v => %v\n\n", key, state)

		_, _, instanceId, _, _, err := core.ParseUnit(key)
		if err != nil {
			return nil, err
		}

		instance := model.MapToInstance(instanceId, &state, nil)

		instances = append(instances, instance)
	}

	//fmt.Printf("%v", status)

	return instances, nil
}
