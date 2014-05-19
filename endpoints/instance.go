package endpoints

import (
	"net/http"

	"github.com/jxaas/jxaas/core"
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
		self.instance = huddle.NewInstance(self.Parent.Parent.Parent.Tenant, self.Parent.BundleType, self.InstanceId)
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

func (self *EndpointInstance) HttpGet() (*model.Instance, error) {
	model, err := self.getInstance().GetState()
	if err == nil && model == nil {
		return nil, rs.ErrNotFound()
	}
	return model, err
}

func (self *EndpointInstance) HttpPut(request *model.Instance) (*model.Instance, error) {
	err := self.getInstance().Configure(request)
	if err != nil {
		return nil, err
	}

	return self.HttpGet()
}

func (self *EndpointInstance) HttpDelete() (*rs.HttpResponse, error) {
	err := self.getInstance().Delete()
	if err != nil {
		return nil, err
	}

	// TODO: Wait for deletion
	// TODO: Remove machines
	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}

type EndpointHealth struct {
	Parent *EndpointInstance
}

func (self *EndpointInstance) ItemHealth() *EndpointHealth {
	child := &EndpointHealth{}
	child.Parent = self
	return child
}

func (self *EndpointHealth) HttpGet() (*model.Health, error) {
	instance := self.Parent.getInstance()
	repair := false

	health, err := instance.RunHealthCheck(repair)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, rs.ErrNotFound()
	}
	return health, nil
}

func (self *EndpointHealth) HttpPost() (*model.Health, error) {
	instance := self.Parent.getInstance()
	repair := true

	health, err := instance.RunHealthCheck(repair)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, rs.ErrNotFound()
	}
	return health, nil
}
