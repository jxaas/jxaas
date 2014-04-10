package endpoints

import (
	"github.com/jxaas/jxaas/juju"
	"github.com/justinsb/gova/log"
)

type EndpointLog struct {
	Parent *EndpointService
}

type Lines struct {
	Line []string
}

func (self *EndpointLog) HttpGet() (*Lines, error) {
	service := self.Parent.ServiceName()

	// TODO: Inject
	logStore := &juju.JujuLogStore{}
	logStore.BaseDir = "/var/log/juju-justinsb-local/"

	// TODO: Expose units?
	unitId := 0

	logfile, err := logStore.ReadLog(service, unitId)
	if err != nil {
		log.Warn("Error reading log: %v", unitId, err)
		return nil, err
	}
	if logfile == nil {
		log.Warn("Log not found: %v", unitId)
		return nil, nil
	}

	lines := &Lines{}
	lines.Line = make([]string, 0)

	logfile.ReadLines(func(line string) (bool, error) {
		lines.Line = append(lines.Line, line)
		return true, nil
	})

	return lines, nil
}
