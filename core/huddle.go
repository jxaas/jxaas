package core

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/juju/juju/state/api"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/juju"
)

// A Huddle is a group of servers. For us, it is a Juju environment into which multiple tenants are deployed.
// Some services are shared across the huddle.
type Huddle struct {
	// URL for the private API (the stub uses this to call private API functions)
	PrivateUrl string

	System         *System
	SharedServices map[string]*SharedService

	JujuClient *juju.Client

	// A lock for operations that aren't concurrent-safe
	mutex sync.Mutex

	// A map of ports assigned
	// Acts both as a cache and a staging area for ports that we assign, but
	// where the service doesn't yet exist so we can't store the port
	assignedPublicPorts map[string]int
}

func NewHuddle(system *System, bundleStore *bundle.BundleStore, jujuApi *juju.Client, privateUrl string) (*Huddle, error) {
	key := "shared"

	systemBundle, err := bundleStore.GetSystemBundle(key)
	if err != nil {
		log.Warn("Error loading system bundle: %v", key, err)
		return nil, err
	}

	if systemBundle == nil {
		log.Warn("Cannot load system bundle: %v", key, err)
		return nil, nil
	}

	info, err := systemBundle.Deploy(jujuApi)
	if err != nil {
		log.Warn("Error deploying system bundle", err)
		return nil, err
	}

	huddle := &Huddle{}
	huddle.PrivateUrl = privateUrl
	huddle.SharedServices = map[string]*SharedService{}
	huddle.assignedPublicPorts = map[string]int{}

	for key, service := range info.Services {
		sharedService := &SharedService{}
		sharedService.JujuName = key
		sharedService.Key = key

		status := service.Status
		if status != nil {
			for _, unit := range status.Units {
				if unit.PublicAddress != "" {
					sharedService.PublicAddress = unit.PublicAddress
				}
			}
		}

		huddle.SharedServices[key] = sharedService
	}

	huddle.JujuClient = jujuApi
	huddle.System = system
	// TODO: Wait until initialized or offer a separate 'bootstrap' command

	{
		check := &HealthCheckAllInstances{}
		check.huddle = huddle
		check.repair = true
		system.Scheduler.AddTask(check, time.Minute*1)
	}

	{
		scaling := &AutoScaleAllInstances{}
		scaling.huddle = huddle
		system.Scheduler.AddTask(scaling, time.Minute*1)
	}

	{
		task := &CleanupOldMachines{}
		task.huddle = huddle
		system.Scheduler.AddTask(task, time.Minute*5)
	}

	return huddle, nil
}

// Implement fmt.Stringer
func (self *Huddle) String() string {
	return log.AsJson(self)
}

// A Juju service that is used by multiple JXaaS instances
// Used, for example, for logging/monitoring services.
type SharedService struct {
	Key           string
	JujuName      string
	PublicAddress string
}

// Implement fmt.Stringer
func (self *SharedService) String() string {
	return log.AsJson(self)
}

// Returns the URL base for the private API server
func (self *Huddle) GetPrivateUrl() string {
	return self.PrivateUrl
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Returns the IP address of the proxy
func (self *Huddle) getProxyHost() (string, error) {
	proxyServiceKey := "jx-haproxy"
	proxyService := self.SharedServices[proxyServiceKey]
	if proxyService == nil {
		log.Warn("Unable to find proxy service: %v", proxyServiceKey)
		return "", errors.New("Unable to find proxy service")
	}

	return proxyService.PublicAddress, nil
}

// Assigns a public port to the serviceId
func (self *Huddle) assignPublicPort(serviceId string) (int, bool, error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var port int

	port, found := self.assignedPublicPorts[serviceId]
	if found {
		return port, false, nil
	}

	// TODO: Filter?
	prefix := ""
	statuses, err := self.JujuClient.GetServiceStatusList(prefix)
	if err != nil {
		return 0, false, err
	}

	publicPorts := []int{}

	for _, publicPort := range self.assignedPublicPorts {
		publicPorts = append(publicPorts, publicPort)
	}

	for key, _ := range statuses {
		var publicPort int

		publicPort, found := self.assignedPublicPorts[key]
		if found {
			assert.That(contains(publicPorts, publicPort))
			continue
		}

		log.Debug("Looking for public port annotation on: %v", key)

		annotations, err := self.JujuClient.GetServiceAnnotations(key)
		if err != nil {
			return 0, false, err
		}

		publicPortString := annotations[ANNOTATION_KEY_PUBLIC_PORT]
		publicPortString = strings.TrimSpace(publicPortString)
		if publicPortString == "" {
			continue
		}
		publicPort, err = strconv.Atoi(publicPortString)
		if err != nil {
			log.Warn("Error parsing public port on %v: %v", key, publicPortString, err)
		}
		self.assignedPublicPorts[key] = publicPort

		publicPorts = append(publicPorts, publicPort)
	}

	// This approach breaks down if the ports are densely assigned
	if len(publicPorts) > 9000 {
		return 0, false, fmt.Errorf("Too many ports already assigned")
	}

	for {
		port = 10000 + rand.Intn(10000)
		if contains(publicPorts, port) {
			continue
		}

		log.Debug("Public ports already assigned: %v", publicPorts)
		log.Info("Assigned port: %v", port)
		break
	}

	// We can't set the port yet; the service likely doesn't yet exist
	//	err = self.Instance.setPublicPort(port)
	//	if err != nil {
	//		return 0, err
	//	}

	// Instead we set the port in the map; this map is how we avoid double allocations before
	// we've created the service
	self.assignedPublicPorts[serviceId] = port

	return port, true, nil
}

func (self *Huddle) ListAllInstances() ([]*Instance, error) {
	prefix := "u"

	statuses, err := self.JujuClient.GetServiceStatusList(prefix)
	if err != nil {
		return nil, err
	}
	if statuses == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	instances := []*Instance{}
	for key, state := range statuses {
		tenant, bundleTypeId, instanceId, module, _, err := ParseUnit(key)
		if err != nil {
			log.Debug("Ignoring unparseable service: %v", key)
			continue
		}

		bundleType := self.System.GetBundleType(bundleTypeId)
		if bundleType == nil {
			log.Debug("Ignoring unknown bundle type: %v", bundleTypeId)
			continue
		}

		if module != bundleType.PrimaryJujuService() {
			continue
		}

		i := self.NewInstance(tenant, bundleType, instanceId)
		i.cacheState(&state)

		instances = append(instances, i)
	}

	return instances, nil
}

func (self *Huddle) ListInstances(tenant string, bundleType bundletype.BundleType) ([]*Instance, error) {
	prefix := self.jujuPrefix(tenant, bundleType)

	statuses, err := self.JujuClient.GetServiceStatusList(prefix)
	if err != nil {
		return nil, err
	}
	if statuses == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	instances := []*Instance{}
	for key, state := range statuses {
		_, bundleTypeId, instanceId, module, _, err := ParseUnit(key)
		if err != nil {
			log.Debug("Ignoring unparseable service: %v", key)
			continue
		}

		assert.That(bundleTypeId == bundleType.Key())

		if module != bundleType.PrimaryJujuService() {
			continue
		}

		i := self.NewInstance(tenant, bundleType, instanceId)
		i.cacheState(&state)

		instances = append(instances, i)
	}

	return instances, nil
}

func (self *Huddle) jujuPrefix(tenant string, bundleType bundletype.BundleType) string {
	tenant = strings.Replace(tenant, "-", "", -1)

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	prefix := "u" + tenant + "-" + bundleType.Key() + "-"

	return prefix
}

func (self *Huddle) cleanupOldMachines(state map[string]int, threshold int) (map[string]int, error) {
	status, err := self.JujuClient.GetSystemStatus()
	if err != nil {
		log.Warn("Error getting system status", err)
		return nil, err
	}

	unitsByMachine := map[string]*api.UnitStatus{}

	for _, serviceStatus := range status.Services {
		for _, unitStatus := range serviceStatus.Units {
			machineId := unitStatus.Machine
			unitsByMachine[machineId] = &unitStatus
		}
	}

	idleMachines := map[string]*api.MachineStatus{}
	for machineId, machineStatus := range status.Machines {
		unit := unitsByMachine[machineId]
		if unit != nil {
			continue
		}
		idleMachines[machineId] = &machineStatus
	}

	idleCounts := map[string]int{}
	for machineId, _ := range idleMachines {
		idleCount := state[machineId]
		idleCount++
		idleCounts[machineId] = idleCount
	}

	for machineId, idleCount := range idleCounts {
		if idleCount < threshold {
			continue
		}

		if machineId == "0" {
			// Machine id 0 is special (the system machine); we can't destroy it
			continue
		}

		log.Info("Machine is idle; removing: %v", machineId)
		err = self.JujuClient.DestroyMachine(machineId)
		if err != nil {
			log.Warn("Failed to delete machine %v", machineId, err)
		}
	}

	return idleCounts, nil
}
