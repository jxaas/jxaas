package core

import (
	"strconv"
	"strings"
	"time"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/checks"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
)

const (
	ANNOTATION_PREFIX_RELATIONINFO  = "__jxaas_relinfo_"
	RELATIONINFO_METADATA_TIMESTAMP = "timestamp"

	ANNOTATION_PREFIX_SYSTEM = "__jxaas_system_"

	// TODO: Should we just find the public-port annotation on the proxy?
	SYSTEM_KEY_PUBLIC_PORT = "public_port"

	ANNOTATION_KEY_PUBLIC_PORT = ANNOTATION_PREFIX_SYSTEM + SYSTEM_KEY_PUBLIC_PORT
)

// Builds an Instance object representing a particular JXaaS Instance.
// This just builds the object, it does not e.g. check that the instance already exists.
func (self *Huddle) GetInstance(tenant string, bundleType bundletype.BundleType, instanceId string) *Instance {
	s := &Instance{}
	s.huddle = self
	s.tenant = tenant
	s.bundleType = bundleType
	s.instanceId = instanceId

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + bundleType.Key() + "-"

	prefix = prefix + instanceId + "-"

	s.jujuPrefix = prefix

	prefix = prefix + bundleType.PrimaryJujuService()
	s.primaryServiceId = prefix

	return s
}

// A JXaaS instance
type Instance struct {
	huddle     *Huddle
	tenant     string
	bundleType bundletype.BundleType
	instanceId string

	jujuPrefix       string
	primaryServiceId string
}

// Returns the current state of the instance
func (self *Instance) GetState() (*model.Instance, error) {
	client := self.huddle.JujuClient

	primaryServiceId := self.primaryServiceId
	status, err := client.GetServiceStatus(primaryServiceId)

	config, err := client.FindConfig(primaryServiceId)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, nil
	}

	log.Debug("Service state: %v", status)

	instance := model.MapToInstance(self.instanceId, status, config)

	serviceKeys, err := self.getBundleKeys()
	if err != nil {
		return nil, err
	}

	if serviceKeys == nil {
		return nil, rs.ErrNotFound()
	}

	// TODO: This is pretty expensive... we could just check to see if properties have been set
	for serviceId, _ := range serviceKeys {
		if serviceId == primaryServiceId {
			continue
		}

		status, err := client.GetServiceStatus(serviceId)
		if err != nil {
			log.Warn("Error while fetching status of service: %v", serviceId, err)
			instance.Status = "pending"
		} else if status == nil {
			log.Warn("No status for service: %v", serviceId)
			instance.Status = "pending"
		} else {
			log.Info("Got state of secondary service: %v => %v", serviceId, status)
			for _, unitStatus := range status.Units {
				model.MergeInstanceStatus(instance, &unitStatus)
			}
		}
	}

	// TODO: This is a bit of a hack also.  How should we wait for properties to be set?
	annotations, err := client.GetServiceAnnotations(primaryServiceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return nil, err
	}

	log.Info("Annotations on %v: %v", primaryServiceId, annotations)

	// TODO: Only if otherwise ready
	annotationsReady := self.bundleType.IsStarted(annotations)

	if !annotationsReady {
		log.Info("Instance not started (per annotations): %v", annotations)
		instance.Status = "pending"
	}

	log.Info("Status of %v: %v", primaryServiceId, instance.Status)

	return instance, nil
}

// Deletes the instance.
// This deletes all Juju services that make up the instance.
func (self *Instance) Delete() error {
	prefix := self.jujuPrefix
	client := self.huddle.JujuClient

	statuses, err := client.GetServiceStatusList(prefix)
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

// Gets any log entries for the instance
func (self *Instance) GetLog() (*model.LogData, error) {
	service := self.primaryServiceId

	// TODO: Inject
	logStore := &juju.JujuLogStore{}

	client := self.huddle.JujuClient
	logStore.BaseDir = client.GetLogDir()

	// TODO: SSH?

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

// Store the relation properties, as set by a consuming unit.
func (self *Instance) SetRelationInfo(unitId string, relationId string, properties map[string]string) error {
	// Annotations on relations aren't supported, and it is tricky to get the relation id
	// So tag it on the service instead

	pairs := make(map[string]string)
	for k, v := range properties {
		pairs[ANNOTATION_PREFIX_RELATIONINFO+unitId+"_"+relationId+"__"+k] = v
	}
	pairs[ANNOTATION_PREFIX_RELATIONINFO+unitId+"_"+relationId+"_"+RELATIONINFO_METADATA_TIMESTAMP] = strconv.Itoa(time.Now().Second())

	return self.setServiceAnnotations(pairs)
}

// Sets annotations on the specified instance.
func (self *Instance) setServiceAnnotations(pairs map[string]string) error {
	serviceId := self.primaryServiceId

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

// Sets the annotation that stores the public port
func (self *Instance) setPublicPort(port int) error {
	pairs := map[string]string{}
	pairs[ANNOTATION_KEY_PUBLIC_PORT] = strconv.Itoa(port)

	return self.setServiceAnnotations(pairs)
}

// Delete any relation properties relating to the specified unit; that unit is going away.
func (self *Instance) DeleteRelationInfo(unitId string, relationId string) error {
	client := self.huddle.JujuClient

	serviceId := self.primaryServiceId

	prefix := ANNOTATION_PREFIX_RELATIONINFO + unitId + "_" + relationId + "_"

	annotations, err := client.GetServiceAnnotations(serviceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return err
	}

	deleteKeys := []string{}

	for tagName, _ := range annotations {
		if !strings.HasPrefix(tagName, prefix) {
			continue
		}
		deleteKeys = append(deleteKeys, tagName)
	}

	if len(deleteKeys) != 0 {
		log.Info("Deleting annotations on service %v: %v", serviceId, deleteKeys)

		err = client.DeleteServiceAnnotations(serviceId, deleteKeys)
		if err != nil {
			log.Warn("Error deleting annotations", err)
			return err
		}
	}

	return nil
}

// Retrieve the stored relation properties.
func (self *Instance) GetRelationInfo(relationKey string) (*model.RelationInfo, error) {
	serviceId := self.primaryServiceId

	relationInfo := &model.RelationInfo{}
	relationInfo.Properties = make(map[string]string)

	client := self.huddle.JujuClient

	status, err := client.GetServiceStatus(serviceId)
	if err != nil {
		log.Warn("Error while fetching status of service: %v", serviceId, err)
		return nil, err
	}

	relationInfo.PublicAddresses = []string{}

	if status != nil {
		log.Info("unitStatus: %v", log.AsJson(status))
		for _, unitStatus := range status.Units {
			if unitStatus.PublicAddress == "" {
				continue
			}
			relationInfo.PublicAddresses = append(relationInfo.PublicAddresses, unitStatus.PublicAddress)
		}
	} else {
		log.Warn("No status found for service: %v", serviceId)
		return nil, nil
	}

	annotations, err := client.GetServiceAnnotations(serviceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return nil, err
	}

	//log.Debug("Service annotations: %v", annotations)

	systemProperties := map[string]string{}
	relationMetadata := map[string]string{}

	relationProperties := []model.RelationProperty{}

	for tagName, v := range annotations {
		if strings.HasPrefix(tagName, ANNOTATION_PREFIX_SYSTEM) {
			key := tagName[len(ANNOTATION_PREFIX_SYSTEM):]
			systemProperties[key] = v
			continue
		}

		if !strings.HasPrefix(tagName, ANNOTATION_PREFIX_RELATIONINFO) {
			//log.Debug("Prefix mismatch: %v", tagName)
			continue
		}
		suffix := tagName[len(ANNOTATION_PREFIX_RELATIONINFO):]
		tokens := strings.SplitN(suffix, "_", 3)
		if len(tokens) < 3 {
			log.Debug("Ignoring unparseable tag: %v", tagName)
			continue
		}

		unitId := tokens[0]
		relationId := tokens[1]
		key := tokens[2]
		if key[0] != '_' {
			relationMetadata[key] = v
			continue
		}

		relationTokens := strings.SplitN(relationId, ":", 2)
		if len(relationTokens) != 2 {
			log.Debug("Ignoring unparseable relation id in tag: %v", tagName)
			continue
		}

		relationProperty := model.RelationProperty{}
		relationProperty.UnitId = unitId
		assert.That(key[0] == '_')
		relationProperty.Key = key[1:]
		relationProperty.Value = v
		relationProperty.RelationType = relationTokens[0]
		relationProperty.RelationKey = relationTokens[1]
		relationProperties = append(relationProperties, relationProperty)
	}

	builder := &bundletype.RelationBuilder{}
	builder.Relation = relationKey
	builder.Properties = relationProperties

	// TODO: Skip proxy host on EC2?
	useProxyHost := true

	if useProxyHost && systemProperties[SYSTEM_KEY_PUBLIC_PORT] != "" {
		publicPortString := systemProperties[SYSTEM_KEY_PUBLIC_PORT]
		publicPort, err := strconv.Atoi(publicPortString)
		if err != nil {
			log.Warn("Error parsing public port: %v", publicPortString, err)
			return nil, err
		}

		proxyHost, err := self.huddle.getProxyHost()
		if err != nil {
			log.Warn("Error fetching proxy host", err)
			return nil, err
		}

		builder.ProxyHost = proxyHost
		builder.ProxyPort = publicPort
	}

	//	log.Debug("relationProperties: %v", relationProperties)
	//	log.Debug("relationMetadata: %v", relationMetadata)

	relationInfo.Timestamp = relationMetadata[RELATIONINFO_METADATA_TIMESTAMP]
	self.bundleType.BuildRelationInfo(relationInfo, builder)

	return relationInfo, nil
}

func (self *Instance) buildSkeletonTemplateContext() *bundle.TemplateContext {
	huddle := self.huddle

	context := &bundle.TemplateContext{}
	context.SystemServices = map[string]string{}
	for key, service := range huddle.SharedServices {
		context.SystemServices[key] = service.JujuName
	}

	context.SystemImplicits = map[string]string{}
	context.SystemImplicits["jxaas-privateurl"] = huddle.GetPrivateUrl()
	context.SystemImplicits["jxaas-tenant"] = self.tenant
	// TODO: Real credentials here
	context.SystemImplicits["jxaas-user"] = "rpcuser"
	context.SystemImplicits["jxaas-secret"] = "rpcsecret"

	context.PublicPortAssigner = &StubPortAssigner{}

	return context
}

func (self *Instance) getBundle(context *bundle.TemplateContext) (*bundle.Bundle, error) {
	tenant := self.tenant
	bundleType := self.bundleType
	name := self.instanceId

	b, err := bundleType.GetBundle(context, tenant, name)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (self *Instance) getBundleKeys() (map[string]string, error) {
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

	keys := map[string]string{}
	for key, service := range bundle.Services {
		keys[key] = service.Charm
	}
	return keys, nil
}

// Ensures the instance is created and has the specified configuration.
// This method is (supposed to be) idempotent.
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

	publicPortAssigner := &InstancePublicPortAssigner{}
	publicPortAssigner.Instance = self
	context.PublicPortAssigner = publicPortAssigner

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

	if publicPortAssigner.Port != 0 {
		self.setPublicPort(publicPortAssigner.Port)
	}

	return nil
}

// Runs a health check on the instance
func (self *Instance) RunHealthCheck(repair bool) (*model.Health, error) {
	client := self.huddle.JujuClient

	services, err := client.GetServiceStatusList(self.jujuPrefix)
	if err != nil {
		return nil, err
	}

	if services == nil || len(services) == 0 {
		return nil, rs.ErrNotFound()
	}

	health := &model.Health{}
	health.Units = map[string]bool{}

	for serviceId, _ := range services {
		healthChecks := []checks.HealthCheck{}

		if strings.HasSuffix(serviceId, "-mysql") {
			checkService := &checks.ServiceHealthCheck{}
			checkService.ServiceName = "mysql"
			healthChecks = append(healthChecks, checkService)
		}

		// TODO: We can't "juju run" on subordinate charms
		//		charm := self.huddle.getCharmInfo(service.Charm)
		//
		//		if charm.Subordinate {
		//			continue
		//		}

		for _, healthCheck := range healthChecks {
			result, err := healthCheck.Run(client, serviceId, repair)

			if err != nil {
				log.Info("Health check failed on %v", serviceId, err)
				return nil, err
			}

			for k, healthy := range result.Units {
				overall, exists := health.Units[k]
				if !exists {
					overall = true
				}
				health.Units[k] = overall && healthy
			}
		}
	}

	//	for unitJujuId, _ := range health {
	//		_, _, _, _, unitId, err := core.ParseUnit(unitJujuId)
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//}

	return health, nil
}
