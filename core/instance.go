package core

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"launchpad.net/juju-core/state/api"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

const (
	// Used to store the (user-set) instance config
	ANNOTATION_PREFIX_INSTANCECONFIG = "__jxaas_config_"

	// Used to store relation properties
	ANNOTATION_PREFIX_RELATIONINFO  = "__jxaas_relinfo_"
	RELATIONINFO_METADATA_TIMESTAMP = "timestamp"

	// Used to store a few system housekeeping items
	ANNOTATION_PREFIX_SYSTEM = "__jxaas_system_"

	// TODO: Should we just find the public-port annotation on the proxy?
	SYSTEM_KEY_PUBLIC_PORT     = "public_port"
	ANNOTATION_KEY_PUBLIC_PORT = ANNOTATION_PREFIX_SYSTEM + SYSTEM_KEY_PUBLIC_PORT

	SYSTEM_KEY_SCALING_POLICY     = "scaling_policy"
	ANNOTATION_KEY_SCALING_POLICY = ANNOTATION_PREFIX_SYSTEM + SYSTEM_KEY_SCALING_POLICY

	SYSTEM_KEY_LAST_STATE     = "last_state"
	ANNOTATION_KEY_LAST_STATE = ANNOTATION_PREFIX_SYSTEM + SYSTEM_KEY_LAST_STATE

	SYSTEM_KEY_LAST_STATE_TIMESTAMP     = "last_state_timestamp"
	ANNOTATION_KEY_LAST_STATE_TIMESTAMP = ANNOTATION_PREFIX_SYSTEM + SYSTEM_KEY_LAST_STATE_TIMESTAMP

	// Artificial delay before marking a service as started (in seconds)
	DELAY_STARTED = 15
)

// Builds an Instance object representing a particular JXaaS Instance.
// This just builds the object, it does not e.g. check that the instance already exists.
func (self *Huddle) NewInstance(tenant string, bundleType bundletype.BundleType, instanceId string) *Instance {
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

// Implement fmt.Stringer
func (self *Instance) String() string {
	return "Instance::" + self.tenant + ":" + self.bundleType.Key() + ":" + self.instanceId
}

// Returns the current state of the instance
func (self *Instance) GetState() (*model.Instance, error) {
	state, err := self.getState0()

	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, nil
	}

	// We inject an artificial pause of 10 seconds between Juju telling us the service is ready,
	// and us marking the service ready.  This is because we normally have a relation to a load balancer like nginx
	// But there is no way to know whether that relation is ready: juju can't tell us the status of a relation

	// TODO: Get the relation state instead.  This is a crazy hack.
	// TODO: I'm not even sure this is actually needed... it may have just been other problems!
	jujuClient := self.GetJujuClient()
	primaryServiceId := self.primaryServiceId

	lastState := state.SystemProperties[SYSTEM_KEY_LAST_STATE]
	lastStateTimestamp := state.SystemProperties[SYSTEM_KEY_LAST_STATE_TIMESTAMP]

	shouldUpdate := true
	delay := false

	now := time.Now().Unix()

	if state.Model.Status == "started" && lastState != "started" {
		if lastStateTimestamp == "" {
			shouldUpdate = true
			delay = true
		} else {
			t, err := strconv.ParseInt(lastStateTimestamp, 10, 64)
			if err != nil {
				log.Warn("Error parsing ANNOTATION_KEY_LAST_STATE_TIMESTAMP: %v", lastStateTimestamp, err)
				// Bypass the delay...
				t = 0
			}

			if (now - t) < DELAY_STARTED {
				delay = true
				shouldUpdate = false
			} else {
				delay = false
				shouldUpdate = true
			}
		}
	}

	if shouldUpdate {
		pairs := make(map[string]string)
		pairs[ANNOTATION_KEY_LAST_STATE] = state.Model.Status
		pairs[ANNOTATION_KEY_LAST_STATE_TIMESTAMP] = strconv.FormatInt(now, 10)

		jujuClient.SetServiceAnnotations(primaryServiceId, pairs)
	}

	if delay {
		log.Warn("Delaying service ready for %v", primaryServiceId)
		state.Model.Status = "pending"
	}

	return state.Model, nil
}

// Like GetState, but only returns true/false if it exists; may be much faster
func (self *Instance) Exists() (bool, error) {
	// TOOD: Optimize
	info, err := self.GetState()
	if err != nil {
		return false, err
	}
	return (info != nil), nil
}

type instanceState struct {
	Model            *model.Instance
	SystemProperties map[string]string
	RelationMetadata map[string]string
	PublicAddresses  []string

	Relations map[string]map[string]string

	Units map[string]map[string]api.UnitStatus
}

// TODO: Get rid of relationProperty entirely?
type relationProperty struct {
	UnitId string

	RelationType string
	RelationKey  string

	Key   string
	Value string
}

func (self *relationProperty) String() string {
	return log.AsJson(self)
}

// Returns the current state of the instance
func (self *Instance) getState0() (*instanceState, error) {
	jujuClient := self.GetJujuClient()

	primaryServiceId := self.primaryServiceId
	status, err := jujuClient.GetServiceStatus(primaryServiceId)

	// XXX: check err?

	jujuService, err := jujuClient.FindService(primaryServiceId)
	if err != nil {
		return nil, err
	}

	if status == nil {
		log.Warn("No state found for %v", primaryServiceId)
		return nil, nil
	}

	log.Debug("Service state: %v", status)

	state := &instanceState{}
	state.Model = model.MapToInstance(self.instanceId, status, jujuService)

	for k, v := range self.bundleType.GetDefaultOptions() {
		option, found := state.Model.OptionDescriptions[k]
		if !found {
			log.Warn("Option not found in OptionDescriptions %v in %v", k, state.Model.OptionDescriptions)
			continue
		}
		option.Default = v
		state.Model.OptionDescriptions[k] = option
	}

	state.Units = map[string]map[string]api.UnitStatus{}

	state.Units[primaryServiceId] = status.Units

	state.PublicAddresses = []string{}
	for _, unitStatus := range status.Units {
		if unitStatus.PublicAddress == "" {
			continue
		}
		state.PublicAddresses = append(state.PublicAddresses, unitStatus.PublicAddress)
	}

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

		status, err := jujuClient.GetServiceStatus(serviceId)
		if err != nil {
			log.Warn("Error while fetching status of service: %v", serviceId, err)
			state.Model.Status = "pending"
		} else if status == nil {
			log.Warn("No status for service: %v", serviceId)
			state.Model.Status = "pending"
		} else {
			log.Info("Got state of secondary service: %v => %v", serviceId, status)
			for _, unitStatus := range status.Units {
				model.MergeInstanceStatus(state.Model, &unitStatus)
			}
		}

		if status != nil {
			state.Units[serviceId] = status.Units
		}
	}

	// TODO: This is a bit of a hack also.  How should we wait for properties to be set?
	annotations, err := jujuClient.GetServiceAnnotations(primaryServiceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return nil, err
	}

	log.Info("Annotations on %v: %v", primaryServiceId, annotations)

	state.Model.Options = map[string]string{}

	state.SystemProperties = map[string]string{}
	state.RelationMetadata = map[string]string{}

	relationList := []relationProperty{}

	for tagName, v := range annotations {
		if strings.HasPrefix(tagName, ANNOTATION_PREFIX_INSTANCECONFIG) {
			key := tagName[len(ANNOTATION_PREFIX_INSTANCECONFIG):]
			state.Model.Options[key] = v
			continue
		}

		if strings.HasPrefix(tagName, ANNOTATION_PREFIX_SYSTEM) {
			key := tagName[len(ANNOTATION_PREFIX_SYSTEM):]
			state.SystemProperties[key] = v
			continue
		}

		if strings.HasPrefix(tagName, ANNOTATION_PREFIX_RELATIONINFO) {
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
				state.RelationMetadata[key] = v
				continue
			}

			relationTokens := strings.SplitN(relationId, ":", 2)
			if len(relationTokens) != 2 {
				log.Debug("Ignoring unparseable relation id in tag: %v", tagName)
				continue
			}

			relationProperty := relationProperty{}
			relationProperty.UnitId = unitId
			assert.That(key[0] == '_')
			relationProperty.Key = key[1:]
			relationProperty.Value = v
			relationProperty.RelationType = relationTokens[0]
			relationProperty.RelationKey = relationTokens[1]
			relationList = append(relationList, relationProperty)

			continue
		}
	}

	state.Relations = map[string]map[string]string{}
	for _, relation := range relationList {
		relationType := relation.RelationType
		relations, found := state.Relations[relationType]
		if !found {
			relations = map[string]string{}
			state.Relations[relationType] = relations
		}
		relations[relation.Key] = relation.Value
	}

	// TODO: Only if otherwise ready?
	annotationsReady := self.bundleType.IsStarted(state.Relations)

	// For a subordinate charm service (e.g. multimysql), we just watch for the annotation
	if annotationsReady && state.Model.Status == "" && len(status.SubordinateTo) != 0 {
		log.Info("Subordinate instance started (per annotations): %v", self)
		state.Model.Status = "started"
	}

	if !annotationsReady {
		log.Info("Instance not started (per annotations): %v", state.Relations)
		state.Model.Status = "pending"
	}

	log.Info("Status of %v: %v", primaryServiceId, state.Model.Status)

	// TODO: Fetch inherited properties from primary service and merge

	return state, nil
}

// Deletes the instance.
// This deletes all Juju services that make up the instance.
func (self *Instance) Delete() error {
	jujuClient := self.GetJujuClient()
	prefix := self.jujuPrefix

	statuses, err := jujuClient.GetServiceStatusList(prefix)
	if err != nil {
		return err
	}
	for serviceId, _ := range statuses {
		log.Debug("Destroying service %v", serviceId)

		err = jujuClient.ServiceDestroy(serviceId)
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
	jujuClient := self.GetJujuClient()
	service := self.primaryServiceId

	logStore, err := jujuClient.GetLogStore()
	if err != nil {
		log.Warn("Error fetching Juju log store", err)
		return nil, err
	}

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
	pairs[ANNOTATION_PREFIX_RELATIONINFO+unitId+"_"+relationId+"_"+RELATIONINFO_METADATA_TIMESTAMP] = strconv.FormatInt(time.Now().Unix(), 10)

	return self.setServiceAnnotations(pairs)
}

// Store the instance configuration, as set by the user
func (self *Instance) setInstanceConfig(properties map[string]string) error {
	pairs := make(map[string]string)
	for k, v := range properties {
		pairs[ANNOTATION_PREFIX_INSTANCECONFIG+k] = v
	}
	return self.setServiceAnnotations(pairs)
}

// Sets annotations on the specified instance.
func (self *Instance) setServiceAnnotations(pairs map[string]string) error {
	jujuClient := self.GetJujuClient()
	serviceId := self.primaryServiceId

	log.Info("Setting annotations on service %v: %v", serviceId, pairs)

	err := jujuClient.SetServiceAnnotations(serviceId, pairs)
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
	jujuClient := self.GetJujuClient()

	serviceId := self.primaryServiceId

	prefix := ANNOTATION_PREFIX_RELATIONINFO + unitId + "_" + relationId + "_"

	annotations, err := jujuClient.GetServiceAnnotations(serviceId)
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

		err = jujuClient.DeleteServiceAnnotations(serviceId, deleteKeys)
		if err != nil {
			log.Warn("Error deleting annotations", err)
			return err
		}
	}

	return nil
}

// Retrieve the relation properties.
// It doesn't seem to be possible to retrieve these direct from Juju,
// so the stubclient stores them for us.
func (self *Instance) GetRelationInfo(relationKey string) (*bundle.Bundle, *model.RelationInfo, error) {
	serviceId := self.primaryServiceId

	// Can we rationalize all this?  We repeat a lot of calls right now...
	state, err := self.getState0()
	if err != nil {
		log.Warn("Error getting instance state", err)
		return nil, nil, err
	}

	if state == nil {
		log.Warn("No status found for service: %v", serviceId)
		return nil, nil, nil
	}

	//	log.Debug("relationProperties: %v", relationProperties)
	//	log.Debug("relationMetadata: %v", relationMetadata)

	context, err := self.buildCurrentTemplateContext(state)
	if err != nil {
		return nil, nil, err
	}
	bundle, err := self.getBundle(context)
	if err != nil {
		return nil, nil, err
	}

	relationInfo, err := self.bundleType.BuildRelationInfo(context, bundle, relationKey)
	if err != nil {
		return nil, nil, err
	}

	if relationInfo != nil {
		if relationInfo.PublicAddresses == nil {
			relationInfo.PublicAddresses = state.PublicAddresses
		}

		relationInfo.Timestamp = state.RelationMetadata[RELATIONINFO_METADATA_TIMESTAMP]
	}

	return bundle, relationInfo, nil
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

	context.PublicPortAssigner = &bundle.StubPortAssigner{}

	return context
}

func (self *Instance) buildCurrentTemplateContext(state *instanceState) (*bundle.TemplateContext, error) {
	var err error

	if state == nil {
		state, err = self.getState0()
		if err != nil {
			log.Warn("Error getting instance state", err)
			return nil, err
		}
	}

	context := self.buildSkeletonTemplateContext()

	// TODO: Need to determine current # of units
	context.NumberUnits = 1

	if state != nil && state.Model != nil {
		context.Options = state.Model.Options
	} else {
		context.Options = map[string]string{}
	}

	publicPortAssigner := &InstancePublicPortAssigner{}
	publicPortAssigner.Instance = self
	context.PublicPortAssigner = publicPortAssigner

	// Populate relation info
	if state != nil {
		context.Relations = state.Relations
	}

	// Populate proxy
	// TODO: Skip proxy host on EC2?
	useProxyHost := true

	if state != nil {
		systemProperties := state.SystemProperties

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

			context.Proxy = &bundle.ProxySettings{}
			context.Proxy.Host = proxyHost
			context.Proxy.Port = publicPort
		}
	}

	return context, nil
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
	var err error

	jujuClient := self.GetJujuClient()

	// Sanitize
	request.Id = ""
	request.Units = nil

	// Record the (requested) configuration options
	instanceConfigChanges := request.Options

	// Get the existing configuration
	context, err := self.buildCurrentTemplateContext(nil)
	if err != nil {
		return err
	}

	// Merge the new configuration options with the existing ones
	if request.NumberUnits != nil {
		context.NumberUnits = *request.NumberUnits
	}
	for k, v := range request.Options {
		context.Options[k] = v
	}

	// Create a bundle from the new configuration
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

	// Save changed config
	if instanceConfigChanges != nil {
		self.setInstanceConfig(instanceConfigChanges)
	}

	// TODO: Is this idempotent?
	publicPortAssigner := context.PublicPortAssigner
	port, assigned := publicPortAssigner.GetAssignedPort()
	if assigned {
		self.setPublicPort(port)
	}

	return nil
}

func (self *Instance) getCurrentBundle(state *instanceState) (*bundle.Bundle, error) {
	context, err := self.buildCurrentTemplateContext(state)
	if err != nil {
		return nil, err
	}

	b, err := self.getBundle(context)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Gets the Juju client
// TODO: Should we expose this?
func (self *Instance) GetJujuClient() *juju.Client {
	jujuClient := self.huddle.JujuClient
	return jujuClient
}

// Runs a health check on the instance
func (self *Instance) RunHealthCheck(repair bool) (*model.Health, error) {
	jujuClient := self.GetJujuClient()

	state, err := self.getState0()
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, rs.ErrNotFound()
	}

	if state.Model == nil {
		log.Debug("No model for %v", self)
		return nil, rs.ErrNotFound()
	}

	if state.Model.Status != "started" {
		log.Info("Skipping health check on not-yet started instance (state %v): %s", state.Model.Status, self)
		return nil, nil
	}

	services, err := jujuClient.GetServiceStatusList(self.jujuPrefix)
	if err != nil {
		return nil, err
	}

	if services == nil || len(services) == 0 {
		return nil, rs.ErrNotFound()
	}

	bundle, err := self.getCurrentBundle(state)
	if err != nil {
		return nil, err
	}

	health := &model.Health{}
	health.Units = map[string]bool{}

	healthChecks, err := self.bundleType.GetHealthChecks(bundle)
	if err != nil {
		return nil, err
	}

	// TODO: We can't "juju run" on subordinate charms
	//		charm := self.huddle.getCharmInfo(service.Charm)
	//
	//		if charm.Subordinate {
	//			continue
	//		}

	for healthCheckId, healthCheck := range healthChecks {
		result, err := healthCheck.Run(self, services, repair)

		if err != nil {
			log.Info("Health check %v failed", healthCheckId, err)
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

	return health, nil
}

func (self *Instance) getScalingPolicy() (*model.ScalingPolicy, error) {
	jujuClient := self.GetJujuClient()
	primaryServiceId := self.primaryServiceId

	annotations, err := jujuClient.GetServiceAnnotations(primaryServiceId)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Ignore?
		return nil, err
	}

	var policy *model.ScalingPolicy
	scalingPolicyJson := annotations[ANNOTATION_KEY_SCALING_POLICY]
	if scalingPolicyJson == "" {
		policy = self.bundleType.GetDefaultScalingPolicy()
	} else {
		policy = &model.ScalingPolicy{}
		err = json.Unmarshal([]byte(scalingPolicyJson), policy)
		if err != nil {
			log.Warn("Error deserializing scaling policy (%v)", scalingPolicyJson, err)
			// TODO: Ignore / go with default?
			return nil, err
		}
	}

	return policy, nil
}

func (self *Instance) setScalingPolicy(policy *model.ScalingPolicy) (*model.ScalingPolicy, error) {
	policyJson, err := json.Marshal(policy)
	if err != nil {
		log.Warn("Error serializing scaling policy", err)
		return nil, err
	}

	pairs := map[string]string{}
	pairs[ANNOTATION_KEY_SCALING_POLICY] = string(policyJson)

	err = self.setServiceAnnotations(pairs)
	if err != nil {
		log.Warn("Error saving scaling policy", err)
		return nil, err
	}

	return policy, nil
}

func (self *Instance) UpdateScalingPolicy(updatePolicy *model.ScalingPolicy) (*model.ScalingPolicy, error) {
	policy, err := self.getScalingPolicy()
	if err != nil {
		return nil, err
	}

	assert.That(updatePolicy != nil)
	if updatePolicy.MetricMin != nil {
		policy.MetricMin = updatePolicy.MetricMin
	}
	if updatePolicy.MetricMax != nil {
		policy.MetricMax = updatePolicy.MetricMax
	}
	if updatePolicy.ScaleMin != nil {
		policy.ScaleMin = updatePolicy.ScaleMin
	}
	if updatePolicy.ScaleMax != nil {
		policy.ScaleMax = updatePolicy.ScaleMax
	}
	if updatePolicy.MetricName != nil {
		policy.MetricName = updatePolicy.MetricName
	}
	if updatePolicy.Window != nil {
		policy.Window = updatePolicy.Window
	}
	return self.setScalingPolicy(policy)
}

// Runs a scaling query and/or change on the instance
func (self *Instance) RunScaling(changeScale bool) (*model.Scaling, error) {
	health := &model.Scaling{}

	instanceState, err := self.GetState()
	if err != nil {
		log.Warn("Error getting instance state", err)
		return nil, err
	}

	assert.That(instanceState.NumberUnits != nil)
	scaleCurrent := *instanceState.NumberUnits
	health.ScaleCurrent = scaleCurrent
	health.ScaleTarget = scaleCurrent

	policy, err := self.getScalingPolicy()
	if err != nil {
		log.Warn("Error fetching scaling policy", err)
		return nil, err
	}

	health.Policy = *policy

	var scaleTarget int

	if policy.MetricName != nil {
		// XXX: Filter by time window
		metricData, err := self.GetMetricValues(*policy.MetricName)
		if err != nil {
			log.Warn("Error retrieving metrics for scaling", err)
			return nil, err
		}

		window := 300
		if policy.Window != nil {
			window = *policy.Window
		}
		duration := time.Duration(-window) * time.Second

		now := time.Now()
		maxTime := now.Unix()
		minTime := now.Add(duration).Unix()

		matches := &model.MetricDataset{}
		for _, point := range metricData.Points {
			t := point.T

			if t < minTime {
				continue
			}

			if t > maxTime {
				continue
			}

			matches.Points = append(matches.Points, point)
		}

		matches.SortPointsByTime()

		lastTime := minTime
		var total float64
		for _, point := range matches.Points {
			t := point.T

			assert.That(t >= lastTime)

			total += float64(float32(t-lastTime) * point.V)

			lastTime = t
		}

		metricCurrent := float32(total / float64(lastTime-minTime))
		log.Info("Average of metric: %v", metricCurrent)

		health.MetricCurrent = metricCurrent

		// TODO: Smart 'target-based' scaling
		scaleDelta := 0
		if policy.MetricMin != nil && metricCurrent < *policy.MetricMin {
			scaleDelta = -1
		} else if policy.MetricMax != nil && metricCurrent > *policy.MetricMax {
			scaleDelta = +1
		}

		scaleTarget = scaleCurrent + scaleDelta
	} else {
		scaleTarget = scaleCurrent
	}

	if policy.ScaleMax != nil && scaleTarget > *policy.ScaleMax {
		scaleTarget = *policy.ScaleMax
	} else if policy.ScaleMin != nil && scaleTarget < *policy.ScaleMin {
		scaleTarget = *policy.ScaleMin
	}

	health.ScaleTarget = scaleTarget

	if changeScale && health.ScaleTarget != scaleCurrent {
		log.Info("Changing scale from %v to %v for %v", scaleCurrent, health.ScaleTarget, self)

		rescale := &model.Instance{}
		rescale.NumberUnits = new(int)
		*rescale.NumberUnits = health.ScaleTarget

		err := self.Configure(rescale)
		if err != nil {
			log.Warn("Error changing scale", err)
			return nil, err
		}
	}

	return health, nil
}

func (self *Instance) cacheState(state *api.ServiceStatus) {
	// TODO: Implement!
}
