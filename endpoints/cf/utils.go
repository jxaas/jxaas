package cf

import (
	"time"

	"github.com/justinsb/gova/log"

	"github.com/jxaas/jxaas/core"
)

func waitReady(instance *core.Instance, timeout int) (bool, error) {
	ready := false
	for i := 0; i < timeout; i++ {
		state, err := instance.GetState()
		if err != nil {
			log.Warn("Error while waiting for instance to become ready", err)
			return false, err
		}

		if state == nil {
			log.Warn("Instance not yet created")
			continue
		}
		status := state.Status

		if status == "started" {
			ready = true
			break
		}

		time.Sleep(time.Second)
		if status == "pending" {
			log.Debug("Instance not ready; waiting", err)
		} else {
			log.Warn("Unknown instance status: %v", status)
		}
	}

	return ready, nil
}
