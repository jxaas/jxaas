package bundle

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"

	"github.com/juju/juju/state/api"
	"launchpad.net/goyaml"

	"github.com/justinsb/gova/log"
)

type DeployInfo struct {
	Services map[string]*DeployServiceInfo
}

type DeployServiceInfo struct {
	Status *api.ServiceStatus
}

func makeConfigYaml(serviceName string, config map[string]string) (string, error) {
	yaml := make(map[string]map[string]string)
	yaml[serviceName] = make(map[string]string)

	for k, v := range config {
		yaml[serviceName][k] = v
	}

	bytes, err := goyaml.Marshal(yaml)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (self *Bundle) Deploy(apiclient *juju.Client) (*DeployInfo, error) {
	log.Debug("Deploying bundle: %v", self)

	info := &DeployInfo{}
	info.Services = map[string]*DeployServiceInfo{}

	for key, service := range self.Services {
		serviceInfo, err := service.deploy(key, apiclient)
		if err != nil {
			return nil, err
		}
		info.Services[key] = serviceInfo
	}

	for _, relation := range self.Relations {
		err := relation.deploy(apiclient)
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

func (self *RelationConfig) deploy(apiclient *juju.Client) error {
	_, err := apiclient.PutRelation(self.From, self.To)
	if err != nil {
		return err
	}
	return nil
}

func (self *ServiceConfig) deploy(jujuServiceId string, apiclient *juju.Client) (*DeployServiceInfo, error) {
	serviceInfo := &DeployServiceInfo{}

	jujuService, err := apiclient.FindService(jujuServiceId)
	if err != nil {
		return nil, err
	}

	charmUrl := self.Charm

	charmInfo, err := apiclient.CharmInfo(charmUrl)
	if err != nil {
		log.Warn("Error reading charm: %v", charmUrl, err)
	}
	if charmInfo == nil {
		log.Warn("Unable to find charm: %v", charmUrl)
	}

	charmUrl = charmInfo.URL

	if jujuService == nil {
		// Create new service

		numUnits := self.NumberUnits

		if charmInfo.Meta.Subordinate {
			numUnits = -1
		}

		configYaml, err := makeConfigYaml(jujuServiceId, self.Options)
		if err != nil {
			return nil, err
		}

		log.Debug("Deploying with YAML: %v", configYaml)

		err = apiclient.ServiceDeploy(
			charmUrl,
			jujuServiceId,
			numUnits,
			configYaml)

		if err != nil {
			return nil, err
		}

		//		for retry := 0; retry < 5; retry++ {
		//			status, err := apiclient.GetStatus(jujuServiceId)
		//			if err != nil {
		//				return err
		//			}
		//			if status != nil {
		//				break
		//			}
		//			log.Info("Service was not yet visible; waiting")
		//			time.Sleep(1 * time.Second)
		//		}
	} else {
		existingInstance := model.MapToInstance(jujuServiceId, nil, jujuService)
		existingServiceOptions := existingInstance.Options
		mergedServiceOptions := map[string]string{}
		{
			for key, value := range existingServiceOptions {
				mergedServiceOptions[key] = value
			}
			for key, value := range self.Options {
				mergedServiceOptions[key] = value
			}
		}

		if !reflect.DeepEqual(existingServiceOptions, mergedServiceOptions) {
			err = apiclient.SetConfig(jujuServiceId, mergedServiceOptions)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug("Configuration unchanged; won't reconfigure")
		}
	}

	if !charmInfo.Meta.Subordinate { // && self.Exposed != nil {
		status, err := apiclient.GetServiceStatus(jujuServiceId)
		if err != nil {
			return nil, err
		}
		if status == nil {
			return nil, fmt.Errorf("Service not found: %v", jujuServiceId)
		}

		serviceInfo.Status = status

		if status.Exposed != self.Exposed {
			err = apiclient.SetExposed(jujuServiceId, self.Exposed)
			if err != nil {
				log.Warn("Error setting service to Exposed=%v", self.Exposed, err)
				return nil, err
			}
		}

		actualUnits := len(status.Units)
		wantUnits := self.NumberUnits
		if actualUnits != wantUnits {
			if actualUnits < wantUnits {
				_, err = apiclient.AddServiceUnits(jujuServiceId, wantUnits-actualUnits)
				if err != nil {
					log.Warn("Error adding units", err)
				}
			} else {
				keys := []string{}
				for key, _ := range status.Units {
					keys = append(keys, key)
				}

				sort.Strings(keys)

				// TODO: Be more intelligent about which unit to kill?
				victims := keys[wantUnits:len(keys)]

				for _, victim := range victims {
					slash := strings.Index(victim, "/")
					unitId, err := strconv.Atoi(victim[slash+1:])
					if err != nil {
						log.Warn("Error parsing UnitId: %v", victim)
						return nil, err
					}

					err = apiclient.DestroyUnit(jujuServiceId, unitId)
					if err != nil {
						log.Warn("Error removing unit: %v/%v", jujuServiceId, unitId, err)
						return nil, err
					}
				}
			}
		}
	}

	//	for _, openPort := range self.OpenPorts {
	//	apiclient.Run(jujuServiceId, nil, ["open-port", openPort])
	//
	////		err = apiclient.OpenPort(jujuServiceId, openPort)
	//		if err != nil {
	//			log.Warn("Error opening port: %v/%v", jujuServiceId, openPort, err)
	//			return nil, err
	//		}
	//	}

	return serviceInfo, nil
}
