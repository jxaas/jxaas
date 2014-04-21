package checks

import (
	"strings"
	"time"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type ServiceHealthCheck struct {
	ServiceName string
}

func (self *ServiceHealthCheck) Run(client *juju.Client, serviceId string, repair bool) (*model.Health, error) {
	command := "service mysql status"
	log.Info("Running command on %v: %v", serviceId, command)

	runResults, err := client.Run(serviceId, command, 5*time.Second)
	if err != nil {
		return nil, err
	}

	health := &model.Health{}
	health.Units = map[string]bool{}

	for _, runResult := range runResults {
		unitJujuId := runResult.UnitId

		code := runResult.Code
		stdout := string(runResult.Stdout)
		stderr := string(runResult.Stderr)

		log.Debug("Result: %v %v %v %v", unitJujuId, code, stdout, stderr)

		healthy := true
		if !strings.Contains(stdout, "start/running") {
			log.Info("Service %v not running on %v", serviceId, unitJujuId)
			healthy = false
		}

		health.Units[unitJujuId] = healthy
	}

	return health, nil
}
