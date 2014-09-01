package endpoints

import (
	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/model"
)

type EndpointInstanceHealth struct {
	Parent *EndpointInstance
}

func (self *EndpointInstanceHealth) HttpGet() (*model.HealthData, error) {
	instance := self.Parent.getInstance()

	// TODO: Use state stored by scheduled health check, rather than running directly?
	health, err := instance.RunHealthCheck(self.repair)
	if err != nil {
		log.Warn("Error running health check on %v", instance, err)
		return nil, err
	}

	log.Debug("Health of %v: %v", instance, health)

	return health, nil
}
