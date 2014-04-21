package endpoints

import "github.com/jxaas/jxaas/model"

type EndpointInstanceMetrics struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceMetrics) HttpGet() (*model.Metrics, error) {
	instance := self.Parent.getInstance()

	return instance.GetMetrics()
}
