package core

import (
	"fmt"
	"strings"
)

func parseService(s string) (tenant, bundleType, instanceId, module string, err error) {
	tokens := strings.SplitN(s, "-", 3)

	if len(tokens) != 3 {
		err = fmt.Errorf("Cannot parse service: %v", s)
		return
	}

	if !strings.HasPrefix(tokens[0], "u") {
		err = fmt.Errorf("Cannot parse tenant")
		return
	}

	tenant = tokens[0][1:]
	bundleType = tokens[1]

	tail := tokens[2]
	lastDash := strings.LastIndex(tail, "-")
	if lastDash == -1 {
		instanceId = tail
		module = ""
	} else {
		instanceId = tail[:lastDash]
		module = tail[(lastDash+1):]
	}

	return
}

func ParseUnit(s string) (tenant, bundleType, instanceId, module, unitId string, err error) {
	lastSlash := strings.LastIndex(s, "/")

	var serviceSpec string
	if lastSlash != -1 {
		unitId = s[lastSlash+1:]
		serviceSpec = s[:lastSlash]
	} else {
		unitId = ""
		serviceSpec = s
	}

	tenant, bundleType, instanceId, module, err = parseService(serviceSpec)
	return
}
