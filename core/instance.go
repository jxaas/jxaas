package core

import (
	"strconv"
	"strings"
	"time"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"

	"github.com/justinsb/gova/log"
)

const (
	PREFIX_RELATIONINFO = "__jxaas_relinfo_"
	SYS_TIMESTAMP       = "timestamp"
)

func (self *Huddle) GetInstance(tenant string, bundleType string, instanceId string) *Instance {
	s := &Instance{}
	s.huddle = self
	s.tenant = tenant
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

func (self *Instance) GetState() (*model.Instance, error) {
	client := self.huddle.JujuClient

	primaryServiceId := self.primaryServiceId
	status, err := client.GetStatus(primaryServiceId)

	config, err := client.FindConfig(primaryServiceId)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, nil
	}

	log.Debug("Service state: %v", status)

	instance := model.MapToInstance(primaryServiceId, status, config)

	serviceNames, err := self.getBundleKeys()
	if err != nil {
		return nil, err
	}

	if serviceNames == nil {
		return nil, rs.ErrNotFound()
	}

	// TODO: This is pretty expensive... we could just check to see if properties have been set
	for _, serviceName := range serviceNames {
		if serviceName == primaryServiceId {
			continue
		}

		status, err := client.GetStatus(serviceName)
		if err != nil {
			log.Warn("Error while fetching status of service: %v", serviceName, err)
			instance.Status = "pending"
		} else if status == nil {
			log.Warn("No status for service: %v", serviceName)
			instance.Status = "pending"
		} else {
			log.Info("Got state of secondary service: %v => %v", serviceName, status)
			for _, unitStatus := range status.Units {
				model.MergeInstanceStatus(instance, &unitStatus)
			}
		}
	}

	return instance, nil
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
		pairs[PREFIX_RELATIONINFO+unitId+"_"+relationId+"__"+k] = v
	}
	pairs[PREFIX_RELATIONINFO+unitId+"_"+relationId+"_"+SYS_TIMESTAMP] = strconv.Itoa(time.Now().Second())

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

	sysInfo := map[string]string{}

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

		if key[0] == '_' {
			relationInfo.Properties[key[1:]] = v
		} else {
			sysInfo[key] = v
		}
	}

	relationInfo.Timestamp = sysInfo[SYS_TIMESTAMP]

	return relationInfo, nil
}

func (self *Instance) buildSkeletonTemplateContext() *bundle.TemplateContext {
	huddle := self.huddle

	context := &bundle.TemplateContext{}
	context.SystemServices = map[string]string{}
	for key, service := range huddle.SharedServices {
		context.SystemServices[key] = service.JujuName
	}

	return context
}

func (self *Instance) getBundle(context *bundle.TemplateContext) (*bundle.Bundle, error) {
	bundleStore := self.huddle.System.BundleStore

	tenant := self.tenant
	bundleType := self.bundleType
	name := self.instanceId

	b, err := bundleStore.GetBundle(context, tenant, bundleType, name)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (self *Instance) getBundleKeys() ([]string, error) {
	// TODO: This is easy to optimize... we really don't need to run the full template...

	context := self.buildSkeletonTemplateContext()

	// TODO: Need to determine current # of units
	context.NumberUnits = 1

	// TODO: Do we need the real config?
	context.Options = map[string]string{}

	bundle, err := self.getBundle(context)
	if err != nil {
		return nil, err
	}
	if bundle == nil {
		return nil, nil
	}

	keys := []string{}
	for key, _ := range bundle.Services {
		keys = append(keys, key)
	}
	return keys, nil
}

func (self *Instance) Configure(request *model.Instance) error {
	jujuClient := self.huddle.JujuClient

	// Sanitize
	request.Id = ""
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]string)
	}
	request.ConfigParameters = nil

	context := self.buildSkeletonTemplateContext()

	if request.NumberUnits == nil {
		// TODO: Need to determine current # of units
		context.NumberUnits = 1
	} else {
		context.NumberUnits = *request.NumberUnits
	}

	context.Options = request.Config

	b, err := self.getBundle(context)
	if err != nil {
		return err
	}
	if b == nil {
		return rs.ErrNotFound()
	}

	_, err = b.Deploy(jujuClient)
	if err != nil {
		return err
	}

	return nil
}
