package endpoints

import (
	"github.com/justinsb/gova/rs"
	"github.com/jxaas/jxaas/model"
)

type EndpointInstanceScaling struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceScaling) HttpGet() (*model.Scaling, error) {
	instance := self.Parent.getInstance()

	exists, err := instance.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, rs.ErrNotFound()
	}

	results, err := instance.RunScaling(false, nil)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (self *EndpointInstanceScaling) HttpPost(scaling *model.ScalingPolicy) (*model.Scaling, error) {
	instance := self.Parent.getInstance()

	exists, err := instance.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, rs.ErrNotFound()
	}

	results, err := instance.RunScaling(true, scaling)
	if err != nil {
		return nil, err
	}
	return results, nil
}
