package checks

import (
	"strings"
	"time"

	"launchpad.net/juju-core/state/api"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type ServiceHealthCheck struct {
	ServiceName string
}

func (self *ServiceHealthCheck) Run(instance jxaas.Instance, jujuState map[string]api.ServiceStatus, repair bool) (*model.Health, error) {
	health := &model.Health{}
	health.Units = map[string]bool{}

	for serviceId, _ := range jujuState {
		self.checkService(instance, serviceId, repair, health)
	}

	return health, nil
}

func (self *ServiceHealthCheck) checkService(instance jxaas.Instance, serviceId string, repair bool, dest *model.Health) error {
	client := instance.GetJujuClient()

	command := "service " + self.ServiceName + " status"
	log.Info("Running command on %v: %v", serviceId, command)

	runResults, err := client.Run(serviceId, nil, command, 5*time.Second)
	if err != nil {
		return err
	}

	for _, runResult := range runResults {
		unitId := juju.ParseUnit(runResult.UnitId)

		code := runResult.Code
		stdout := string(runResult.Stdout)
		stderr := string(runResult.Stderr)

		log.Debug("Result: %v %v %v %v", runResult.UnitId, code, stdout, stderr)

		healthy := true
		if !strings.Contains(stdout, "start/running") {
			log.Info("Service %v not running on %v", serviceId, runResult.UnitId)
			healthy = false

			if repair {
				command := "service " + self.ServiceName + " start"
				log.Info("Running command on %v: %v", serviceId, command)

				_, err := client.Run(serviceId, []string{unitId}, command, 5*time.Second)
				if err != nil {
					return err
				}

			}
		}

		dest.Units[unitId] = healthy
	}

	return nil
}
