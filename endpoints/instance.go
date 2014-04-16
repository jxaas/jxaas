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

type EndpointInstance struct {
	Parent     *EndpointBundle
	InstanceId string
	Huddle     *core.Huddle

	instance *core.Instance
}

func (self *EndpointInstance) ItemMetrics() *EndpointInstanceMetrics {
	child := &EndpointInstanceMetrics{}
	child.Parent = self
	return child
}

func (self *EndpointInstance) ItemLog() *EndpointInstanceLog {
	child := &EndpointInstanceLog{}
	child.Parent = self
	return child
}

func (self *EndpointInstance) ItemRelations() *EndpointRelations {
	child := &EndpointRelations{}
	child.Parent = self
	return child
}

func (self *EndpointInstance) getHuddle() *core.Huddle {
	return self.Huddle
}

func (self *EndpointInstance) getInstance() *core.Instance {
	if self.instance == nil {
		huddle := self.getHuddle()
		self.instance = huddle.GetInstance(self.Parent.Parent.Parent.Tenant, self.Parent.BundleType, self.InstanceId)
	}
	return self.instance
}

//func (self *EndpointInstance) jujuPrefix() string {
//	prefix := self.Parent.jujuPrefix()
//
//	name := self.ServiceKey
//	prefix += name + "-"
//
//	return prefix
//}

func (self *EndpointInstance) HttpGet(apiclient *juju.Client) (*model.Instance, error) {
	model, err := self.getInstance().GetState()
	if err == nil && model == nil {
		return nil, rs.ErrNotFound()
	}
	return model, err
}

func (self *EndpointInstance) HttpPut(apiclient *juju.Client, bundleStore *bundle.BundleStore, huddle *core.Huddle, request *model.Instance) (*model.Instance, error) {
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
	bundleType := self.Parent.BundleType
	name := self.InstanceId

	b, err := bundleStore.GetBundle(context, tenant, bundleType, name)
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

func (self *EndpointInstance) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
	err := self.getInstance().Delete()
	if err != nil {
		return nil, err
	}

	// TODO: Wait for deletion
	// TODO: Remove machines
	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}
