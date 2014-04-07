package juju

import (
	"path"
	"strconv"

	"github.com/justinsb/gova/files"
)

// For now, this uses direct filesystem access (and presumes it is local)
// TODO: Move to SSH
type JujuLogStore struct {
	BaseDir string
}

type JujuLog struct {
	path string
}

func (self *JujuLogStore) ReadLog(service string, unitId int) (*JujuLog, error) {
	// TODO: Block path traversal
	filename := "unit-" + service + "-" + strconv.Itoa(unitId) + ".log"
	path := path.Join(self.BaseDir, filename)
	ok, err := files.Exists(path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	log := &JujuLog{}
	log.path = path
	return log, nil
}

type fileLineIterator struct {
	path string
}

func (self *JujuLog) ReadLines(processor files.FnLineProcessor) error {
	return files.ReadLines(self.path, processor)
}
