package endpoints

import "github.com/jxaas/jxaas/model"

type EndpointInstanceLog struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceLog) HttpGet() (*model.LogData, error) {
	instance := self.Parent.getInstance()

	data, err := instance.GetLog()

	if err != nil {
		return nil, err
	}

	return data, nil
}
