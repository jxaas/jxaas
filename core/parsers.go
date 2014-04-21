package core

import (
	"fmt"
	"strings"
)

func parseService(s string) (tenant, serviceType, instanceId, module string, err error) {
	tokens := strings.SplitN(s, "-", 3)

	if len(tokens) != 3 {
		err = fmt.Errorf("Cannot parse service")
		return
	}

	if !strings.HasPrefix(tokens[0], "u") {
		err = fmt.Errorf("Cannot parse tenant")
		return
	}

	tenant = tokens[0][1:]
	serviceType = tokens[1]

	tail := tokens[2]
	lastDash := strings.LastIndex(tail, "-")
	if lastDash == -1 {
		instanceId = tail
		module = ""
	} else {
		instanceId = tail[:lastDash]
		module = tail[lastDash:]
	}

	return
}

func ParseUnit(s string) (tenant, serviceType, instanceId, module, unitId string, err error) {
	lastSlash := strings.LastIndex(s, "/")

	var serviceSpec string
	if lastSlash != -1 {
		unitId = s[lastSlash+1:]
		serviceSpec = s[:lastSlash]
	} else {
		unitId = ""
		serviceSpec = s
	}

	tenant, serviceType, instanceId, module, err = parseService(serviceSpec)
	return
}
