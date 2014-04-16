package endpoints

import (
	"net/http"
	"strings"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"
)

type EndpointService struct {
	Parent     *EndpointCharm
	InstanceId string

	Huddle *core.Huddle

	instance *core.Instance
}

func (self *EndpointService) ItemMetrics() *EndpointMetrics {
	child := &EndpointMetrics{}
	child.Parent = self
	return child
}

func (self *EndpointService) ItemLog() *EndpointLog {
	child := &EndpointLog{}
	child.Parent = self
	return child
}

func (self *EndpointService) ItemRelations() *EndpointRelations {
	child := &EndpointRelations{}
	child.Parent = self
	return child
}

func (self *EndpointService) getHuddle() *core.Huddle {
	return self.Huddle
}

func (self *EndpointService) getInstance() *core.Instance {
	if self.instance == nil {
		huddle := self.getHuddle()
		self.instance = huddle.GetInstance(self.Parent.Parent.Parent.Tenant, self.Parent.ServiceType, self.InstanceId)
	}
	return self.instance
}

//func (self *EndpointService) jujuPrefix() string {
//	prefix := self.Parent.jujuPrefix()
//
//	name := self.ServiceKey
//	prefix += name + "-"
//
//	return prefix
//}

func (self *EndpointService) HttpGet(apiclient *juju.Client) (*model.Instance, error) {
	model, err := self.getInstance().GetState()
	if err == nil && model == nil {
		return nil, rs.ErrNotFound()
	}
	return model, err
}

func (self *EndpointService) HttpPut(apiclient *juju.Client, bundleStore *bundle.BundleStore, huddle *core.Huddle, request *model.Instance) (*model.Instance, error) {
	// Sanitize
	request.Id = ""
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]string)
	}
	request.ConfigParameters = nil

	context := &bundle.TemplateContext{}
	context.SystemServices = map[string]string{}
	for key, service := range huddle.SharedServices {
		context.SystemServices[key] = service.JujuName
	}

	if request.NumberUnits == nil {
		// TODO: Need to determine current # of units
		context.NumberUnits = 1
	} else {
		context.NumberUnits = *request.NumberUnits
	}

	context.Options = request.Config

	tenant := self.Parent.Parent.Parent.Tenant
	tenant = strings.Replace(tenant, "-", "", -1)
	serviceType := self.Parent.ServiceType
	instanceId := self.InstanceId

	b, err := bundleStore.GetBundle(context, tenant, serviceType, instanceId)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, rs.ErrNotFound()
	}

	_, err = b.Deploy(apiclient)
	if err != nil {
		return nil, err
	}

	return self.HttpGet(apiclient)
}

func (self *EndpointService) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
	err := self.getInstance().Delete()
	if err != nil {
		return nil, err
	}

	// TODO: Wait for deletion
	// TODO: Remove machines
	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}
