package core

import (
	"github.com/justinsb/gova/log"
)

type HealthCheckAllInstances struct {
	huddle *Huddle
	repair bool
}

func (self *HealthCheckAllInstances) Run() error {
	instances, err := self.huddle.ListAllInstances()
	if err != nil {
		log.Warn("Error listing instances", err)
		return err
	}

	for _, instance := range instances {
		health, err := instance.RunHealthCheck(self.repair)
		if err != nil {
			log.Warn("Error running health check on instance: %v", instance, err)
			continue
		}

		// TODO: Check health results and mark instances unhealthy??
		log.Debug("Health of %v: %v", instance, health)
	}

	return nil
}
