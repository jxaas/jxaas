package bundle

import (
	"fmt"
	"reflect"

	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"

	"launchpad.net/goyaml"
	"launchpad.net/juju-core/state/api"

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

	config, err := apiclient.FindConfig(jujuServiceId)
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

	if config == nil {
		// Create new service

		//	curl, err := charm.InferURL(c.CharmName, conf.DefaultSeries())
		//	if err != nil {
		//		return err
		//	}
		//	repo, err := charm.InferRepository(curl, ctx.AbsPath(c.RepoPath))
		//	if err != nil {
		//		return err
		//	}
		//
		//	repo = config.SpecializeCharmRepo(repo, conf)
		//
		//	curl, err = addCharmViaAPI(client, ctx, curl, repo)
		//	if err != nil {
		//		return err
		//	}

		numUnits := self.NumberUnits

		if charmInfo.Meta.Subordinate {
			numUnits = -1
		}

		//		serviceName := "service" + strconv.Itoa(rand.Int())

		//	if serviceName == "" {
		//		serviceName = charmInfo.Meta.Name
		//	}

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
		existingValues := model.MapToConfig(config)
		mergedValues := make(map[string]string)
		{
			for key, value := range existingValues {
				mergedValues[key] = value
			}
			for key, value := range self.Options {
				mergedValues[key] = value
			}
		}

		if !reflect.DeepEqual(existingValues, mergedValues) {
			err = apiclient.SetConfig(jujuServiceId, mergedValues)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug("Configuration unchanged; won't reconfigure")
		}
	}

	if !charmInfo.Meta.Subordinate { // && self.Exposed != nil {
		status, err := apiclient.GetStatus(jujuServiceId)
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

		nUnits := len(status.Units)
		if nUnits != self.NumberUnits {
			log.Warn("NumberUnits mismatch")
		}
	}

	return serviceInfo, nil
}
