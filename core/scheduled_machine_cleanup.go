package core

import (
	"github.com/justinsb/gova/log"
)

type CleanupOldMachines struct {
	huddle *Huddle

	state           map[string]int
	deleteThreshold int
}

func (self *CleanupOldMachines) Run() error {
	state, err := self.huddle.cleanupOldMachines(self.state, self.deleteThreshold)
	if err != nil {
		log.Warn("Error cleaning up old machines", err)
		return err
	}
	self.state = state

	return nil
}

func (self *CleanupOldMachines) String() string {
	return "CleanupOldMachines"
}
