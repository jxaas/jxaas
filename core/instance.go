package core

import (
	"strings"

	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"

	"github.com/justinsb/gova/log"
)

func (self *Huddle) GetInstance(tenant string, bundleType string, instanceId string) *Instance {
	s := &Instance{}
	s.huddle = self
	s.tenant = strings.Replace(tenant, "-", "", -1)
	s.bundleType = bundleType
	s.instanceId = instanceId

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + bundleType + "-"

	prefix = prefix + instanceId + "-"

	s.jujuPrefix = prefix

	primaryJujuService := bundleType
	prefix = prefix + primaryJujuService
	s.primaryServiceId = prefix

	return s
}

type Instance struct {
	huddle     *Huddle
	tenant     string
	bundleType string
	instanceId string

	jujuPrefix       string
	primaryServiceId string
}

const (
	PREFIX_RELATIONINFO = "__jxaas_relinfo_"
)

func (self *Instance) GetState() (*model.Instance, error) {
	serviceName := self.primaryServiceId
	client := self.huddle.JujuClient

	status, err := client.GetStatus(serviceName)

	config, err := client.FindConfig(serviceName)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, nil
	}

	log.Debug("Service state: %v", status)

	return model.MapToInstance(serviceName, status, config), nil
}

func (self *Instance) Delete() error {
	prefix := self.jujuPrefix
	client := self.huddle.JujuClient

	statuses, err := client.GetStatusList(prefix)
	if err != nil {
		return err
	}
	for serviceId, _ := range statuses {
		log.Debug("Destroying service %v", serviceId)

		err = client.ServiceDestroy(serviceId)
		if err != nil {
			log.Warn("Error destroying service: %v", serviceId)
			return err
		}
	}
	// TODO: Wait for deletion
	// TODO: Remove machines
	return nil
}

func (self *Instance) GetLog() (*model.LogData, error) {
	service := self.primaryServiceId

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

	data := &model.LogData{}
	data.Lines = make([]string, 0)

	logfile.ReadLines(func(line string) (bool, error) {
		data.Lines = append(data.Lines, line)
		return true, nil
	})

	return data, nil
}

func (self *Instance) SetRelationInfo(unitId string, relationId string, properties map[string]string) error {
	// Annotations on relations aren't supported, and it is tricky to get the relation id
	// So tag it on the service instead

	serviceId := self.primaryServiceId

	pairs := make(map[string]string)
	for k, v := range properties {
		pairs[PREFIX_RELATIONINFO+unitId+"_"+relationId+"_"+k] = v
	}

	log.Info("Setting annotations on service %v: %v", serviceId, pairs)

	client := self.huddle.JujuClient

	err := client.SetServiceAnnotations(serviceId, pairs)
	if err != nil {
		log.Warn("Error setting annotations", err)
		// TODO: Mask error?
		return err
	}

	return nil
}

func (self *Instance) GetRelationInfo(relationKey string) (*model.RelationInfo, error) {
	serviceId := self.primaryServiceId

	client := self.huddle.JujuClient

	annotations, err := client.GetServiceAnnotations(serviceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return nil, err
	}

	relationIdPrefix := relationKey + ":"

	relationInfo := &model.RelationInfo{}
	relationInfo.Properties = make(map[string]string)

	for tagName, v := range annotations {
		if !strings.HasPrefix(tagName, PREFIX_RELATIONINFO) {
			//log.Debug("Prefix mismatch: %v", tagName)
			continue
		}
		suffix := tagName[len(PREFIX_RELATIONINFO):]
		tokens := strings.SplitN(suffix, "_", 3)
		if len(tokens) < 3 {
			log.Debug("Ignoring unparseable tag: %v", tagName)
			continue
		}

		// unitId = tokens[0]
		relationId := tokens[1]
		if !strings.HasPrefix(relationId, relationIdPrefix) {
			//log.Debug("Relation prefix mismatch: %v", relationId)
			continue
		}

		key := tokens[2]
		relationInfo.Properties[key] = v
	}

	return relationInfo, nil
}
