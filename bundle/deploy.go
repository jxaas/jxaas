package bundle

import (
	"fmt"
	"reflect"

	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/model"

	"launchpad.net/goyaml"

	"github.com/justinsb/gova/log"
)

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

func (self *Bundle) Deploy(apiclient *juju.Client) error {
	for key, service := range self.Services {
		err := service.deploy(key, apiclient)
		if err != nil {
			return err
		}
	}

	for _, relation := range self.Relations {
		err := relation.deploy(apiclient)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *RelationConfig) deploy(apiclient *juju.Client) error {
	_, err := apiclient.PutRelation(self.From, self.To)
	if err != nil {
		return err
	}
	return nil
}

func (self *ServiceConfig) deploy(jujuServiceId string, apiclient *juju.Client) error {
	config, err := apiclient.FindConfig(jujuServiceId)
	if err != nil {
		return err
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
			return err
		}

		log.Debug("Deploying with YAML: %v", configYaml)

		err = apiclient.ServiceDeploy(
			charmUrl,
			jujuServiceId,
			numUnits,
			configYaml)

		if err != nil {
			return err
		}
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
				return err
			}
		} else {
			log.Debug("Configuration unchanged; won't reconfigure")
		}
	}

	if !charmInfo.Meta.Subordinate { // && self.Exposed != nil {
		status, err := apiclient.GetStatus(jujuServiceId)
		if err != nil {
			return err
		}
		if status == nil {
			return fmt.Errorf("Service not found: %v", jujuServiceId)
		}

		if status.Exposed != self.Exposed {
			err = apiclient.SetExposed(jujuServiceId, self.Exposed)
			if err != nil {
				log.Warn("Error setting service to Exposed=%v", self.Exposed, err)
				return err
			}
		}
	}

	return nil
}
