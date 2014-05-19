package checks

import (
	"strings"
	"time"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type ServiceHealthCheck struct {
	ServiceName string
}

func (self *ServiceHealthCheck) Run(instance jxaas.Instance, serviceId string, repair bool) (*model.Health, error) {
	client := instance.GetJujuClient()

	command := "service " + self.ServiceName + " status"
	log.Info("Running command on %v: %v", serviceId, command)

	runResults, err := client.Run(serviceId, nil, command, 5*time.Second)
	if err != nil {
		return nil, err
	}

	health := &model.Health{}
	health.Units = map[string]bool{}

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
					return nil, err
				}

			}
		}

		health.Units[unitId] = healthy
	}

	return health, nil
}
