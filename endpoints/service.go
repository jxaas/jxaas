package endpoints

import (
	"net/http"
	"strings"

	"bitbucket.org/jsantabarbara/jxaas/bundle"
	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/model"
	"bitbucket.org/jsantabarbara/jxaas/rs"
	"github.com/justinsb/gova/log"
)

type EndpointService struct {
	Parent     *EndpointCharm
	ServiceKey string
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

func (self *EndpointService) ItemRelation() *EndpointRelations {
	child := &EndpointRelations{}
	child.Parent = self
	return child
}

func (self *EndpointService) ServiceName() string {
	tenant := self.Parent.Parent.Parent.Tenant
	tenant = strings.Replace(tenant, "-", "", -1)

	serviceType := self.Parent.ServiceType

	serviceKey := self.ServiceKey

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + serviceType + "-"

	if strings.HasPrefix(serviceKey, prefix) {
		// If we already include the prefix, don't re-include it
		// TODO: This is not a great idea
		return serviceKey
	} else {
		return prefix + serviceKey
	}
}

func (self *EndpointService) HttpGet(apiclient *juju.Client) (*model.Instance, error) {
	serviceName := self.ServiceName()
	status, err := apiclient.GetStatus(serviceName)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	config, err := apiclient.FindConfig(serviceName)
	if err != nil {
		return nil, err
	}

	log.Debug("Service state: %v", status)

	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil

	return model.MapToInstance(serviceName, status, config), nil
}

func (self *EndpointService) HttpPut(apiclient *juju.Client, request *model.Instance) (*model.Instance, error) {
	// Sanitize
	request.Id = ""
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]string)
	}
	request.ConfigParameters = nil

	context := &bundle.TemplateContext{}
	context.SystemServices = map[string]string{}
	context.SystemServices["elasticsearch"] = "es1"

	tenant := self.Parent.Parent.Parent.Tenant
	serviceType := self.Parent.ServiceType
	name := self.ServiceName()

	b, err := bundle.GetBundle(context, tenant, serviceType, name)
	if err != nil {
		return nil, err
	}

	err = b.Deploy(apiclient)
	if err != nil {
		return nil, err
	}

	return self.HttpGet(apiclient)
}

func (self *EndpointService) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
	serviceName := self.ServiceName()

	err := apiclient.ServiceDestroy(serviceName)
	if err != nil {
		return nil, err
	}

	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}
