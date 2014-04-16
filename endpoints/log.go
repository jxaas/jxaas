package endpoints

import "github.com/jxaas/jxaas/model"

type EndpointLog struct {
	Parent *EndpointService
}

func (self *EndpointLog) HttpGet() (*model.LogData, error) {
	instance := self.Parent.getInstance()

	data, err := instance.GetLog()

	if err != nil {
		return nil, err
	}

	return data, nil
}
