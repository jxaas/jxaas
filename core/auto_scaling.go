package core

import (
	"github.com/justinsb/gova/log"
)

type AutoScaleAllInstances struct {
	huddle *Huddle
}

func (self *AutoScaleAllInstances) Run() error {
	instances, err := self.huddle.ListAllInstances()
	if err != nil {
		log.Warn("Error listing instances", err)
		return err
	}

	for _, instance := range instances {
		scaling, err := instance.RunScaling(true, nil)
		if err != nil {
			log.Warn("Error running scaling on instance: %v", instance, err)
			continue
		}

		// TODO: Record this, so we can return scaling info from last poll through API
		log.Debug("Scaling-state of %v: %v", instance, scaling)
	}

	return nil
}
