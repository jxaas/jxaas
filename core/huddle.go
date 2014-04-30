package core

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/juju"
)

// System is the top-level object for storing system state
type System struct {
	BundleTypes map[string]bundletype.BundleType
}

// Gets the bundle type by key
func (self *System) GetBundleType(key string) bundletype.BundleType {
	return self.BundleTypes[key]
}

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
