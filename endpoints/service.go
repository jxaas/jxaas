package endpoints

import (
	"net/http"
	"reflect"

	"launchpad.net/goyaml"

	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/rs"
	"github.com/justinsb/gova/log"
)

type EndpointService struct {
	Parent    *EndpointCharm
	ServiceId string
}

func (self *EndpointService) ItemMetrics() *EndpointMetrics {
	child := &EndpointMetrics{}
	child.Parent = self
	return child
}

func (self *EndpointService) ItemLog() *EndpointLog {
	child := &EndpointLog{}
	child.Parent = self
	return child
}

func (self *EndpointService) HttpGet(apiclient *juju.Client) (*Instance, error) {
	status, err := apiclient.GetStatus(self.ServiceId)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	config, err := apiclient.FindConfig(self.ServiceId)
	if err != nil {
		return nil, err
	}

	log.Debug("Service state: %v", status)

	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil

	return MapToInstance(self.ServiceId, status, config), nil
}

func makeConfigYaml(request *Instance) (string, error) {
	id := request.Id

	yaml := make(map[string]map[string]string)
	yaml[id] = make(map[string]string)

	for k, v := range request.Config {
		yaml[id][k] = v
	}

	bytes, err := goyaml.Marshal(yaml)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (self *EndpointService) HttpPut(apiclient *juju.Client, request *Instance) (*Instance, error) {
	// Sanitize
	request.Id = self.ServiceId
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]string)
	}

	config, err := apiclient.FindConfig(self.ServiceId)
	if err != nil {
		return nil, err
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

		//
		//	charmInfo, err := client.CharmInfo(curl.String())
		//	if err != nil {
		//		return err
		//	}

		numUnits := 1

		//		serviceName := "service" + strconv.Itoa(rand.Int())

		//	if serviceName == "" {
		//		serviceName = charmInfo.Meta.Name
		//	}

		charmUrl := "cs:precise/mysql-38"

		configYaml, err := makeConfigYaml(request)
		if err != nil {
			return nil, err
		}

		log.Debug("Deploying with YAML: %v", configYaml)

		err = apiclient.ServiceDeploy(
			charmUrl,
			self.ServiceId,
			numUnits,
			configYaml)

		if err != nil {
			return nil, err
		}
	} else {
		existingValues := MapToConfig(config)
		mergedValues := make(map[string]string)
		{
			for key, value := range existingValues {
				mergedValues[key] = value
			}
			for key, value := range request.Config {
				mergedValues[key] = value
			}
		}

		if !reflect.DeepEqual(existingValues, mergedValues) {
			err = apiclient.SetConfig(self.ServiceId, mergedValues)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug("Configuration unchanged; won't reconfigure")
		}
	}

	if request.Exposed != nil {
		status, err := apiclient.GetStatus(self.ServiceId)
		if err != nil {
			return nil, err
		}
		if status.Exposed != *request.Exposed {
			err = apiclient.SetExposed(self.ServiceId, *request.Exposed)
			if err != nil {
				log.Warn("Error setting service to Exposed=%v", *request.Exposed, err)
				return nil, err
			}
		}
	}

	return self.HttpGet(apiclient)
}

func (self *EndpointService) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
	serviceId := self.ServiceId

	err := apiclient.ServiceDestroy(serviceId)
	if err != nil {
		return nil, err
	}

	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}
