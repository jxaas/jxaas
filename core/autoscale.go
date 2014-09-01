package core

import "github.com/jxaas/jxaas/model"

type AutoScaleRule struct {
	MinUnits int
	MaxUnits int

	Metric string

	LowMark  float32
	HighMark float32
}

func (self *AutoScaleRule) GetMetrics() (*model.Metrics, error) {
	return nil, nil
}

//
//,"{\"Uuid\":\"227e251b-9ad8-4b18-b907-11f223d511f1\",\"Timestamp\":\"2014-04-21T03:05:02.093Z\",
//\"Type\":\"LoadAverage\",\"Logger\":\"LoadAverage\",\"Severity\":7,
//\"Payload\":\"0.34 0.41 0.45 2/2033 9036\",\"EnvVersion\":\"\",
//\"Pid\":8432,\"Hostname\":\"u487c4524761c4eed89f56627f07e9227-mysql-j4-proxy-metrics/0\",
//\"ProcessInputName\":\"LoadAverage.stdout\",\"Load1Min\":\"0.34\",\"Load5Min\":\"0.41\",\"Load15Min\":\"0.45\"}",
