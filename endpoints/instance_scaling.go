package endpoints

import (
	"github.com/justinsb/gova/log"
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

	results, err := instance.RunScaling(false)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (self *EndpointInstanceScaling) HttpPut(policyUpdate *model.ScalingPolicy) (*model.Scaling, error) {
	instance := self.Parent.getInstance()

	log.Info("Policy update: %v", policyUpdate)

	exists, err := instance.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, rs.ErrNotFound()
	}

	if policyUpdate != nil {
		_, err := instance.UpdateScalingPolicy(policyUpdate)
		if err != nil {
			log.Warn("Error updating scaling policy", err)
			return nil, err
		}
	}

	results, err := instance.RunScaling(true)
	if err != nil {
		return nil, err
	}
	return results, nil
}
