package endpoints

import (
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"
	"github.com/jxaas/jxaas/model"
)

type EndpointInstanceHealth struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceHealth) HttpGet() (*model.Health, error) {
	instance := self.Parent.getInstance()
	repair := false

	// TODO: Use state stored by scheduled health check, rather than running directly?
	health, err := instance.RunHealthCheck(repair)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, rs.ErrNotFound()
	}

	log.Debug("Health of %v: %v", instance, health)

	return health, nil
}

func (self *EndpointInstanceHealth) HttpPost() (*model.Health, error) {
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
