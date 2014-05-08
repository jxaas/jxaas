package endpoints

import "github.com/jxaas/jxaas/model"

type EndpointInstanceMetrics struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceMetrics) HttpGet() (*model.Metrics, error) {
	instance := self.Parent.getInstance()

	return instance.GetMetricInfo()
}

func (self *EndpointInstanceMetrics) Item(key string) *EndpointInstanceMetricDataset {
	child := &EndpointInstanceMetricDataset{}
	child.Parent = self
	child.Key = key
	return child
}

type EndpointInstanceMetricDataset struct {
	Parent *EndpointInstanceMetrics
	Key    string
}

func (self *EndpointInstanceMetricDataset) HttpGet() (*model.MetricDataset, error) {
	instance := self.Parent.Parent.getInstance()

	return instance.GetMetricValues(self.Key)
}
