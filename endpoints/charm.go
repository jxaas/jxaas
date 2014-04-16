package endpoints

import (
	"net/http"
	"strings"

	"github.com/jxaas/jxaas/inject"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"
)

type EndpointCharm struct {
	Parent      *EndpointServices
	ServiceType string
}

func (self *EndpointCharm) jujuPrefix() string {
	tenant := self.Parent.Parent.Tenant
	tenant = strings.Replace(tenant, "-", "", -1)

	serviceType := self.ServiceType

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + serviceType + "-"

	return prefix
}

func (self *EndpointCharm) Item(key string, injector inject.Injector) *EndpointService {
	child := &EndpointService{}
	child.Parent = self
	child.InstanceId = key

	injector.Inject(&child.Huddle)

	return child
}

func (self *EndpointCharm) HttpGet(apiclient *juju.Client) ([]*model.Instance, error) {
	prefix := self.jujuPrefix()

	statuses, err := apiclient.GetStatusList(prefix)
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
		if !strings.HasSuffix(key, "-"+self.ServiceType) {
			continue
		}

		//fmt.Printf("%v => %v\n\n", key, state)
		instance := model.MapToInstance(key, &state, nil)

		instances = append(instances, instance)
	}

	//fmt.Printf("%v", status)

	return instances, nil
}
