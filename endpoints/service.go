package endpoints

import (
	"net/http"
	"reflect"

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

	config, err := apiclient.GetConfig(self.ServiceId)
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

func (self *EndpointService) HttpPut(apiclient *juju.Client, request *Instance) (*Instance, error) {
	// Sanitize (just in case)
	request.Id = self.ServiceId
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]ConfigValue)
	}

	config, err := apiclient.GetConfig(self.ServiceId)
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

		var configYAML string
		//	if c.Config.Path != "" {
		//		configYAML, err = c.Config.Read(ctx)
		//		if err != nil {
		//			return err
		//		}
		//	}

		charmUrl := "cs:precise/mysql-38"

		err := apiclient.ServiceDeploy(
			charmUrl,
			self.ServiceId,
			numUnits,
			configYAML,
		)

		if err != nil {
			return nil, err
		}
	} else {
		existingConfig := MapToConfiguration(config)

		existingValues := make(map[string]string)
		{
			for key, value := range existingConfig {
				existingValues[key] = value.Value
			}
		}
		mergedValues := make(map[string]string)
		{
			for key, value := range existingConfig {
				mergedValues[key] = value.Value
			}
			for key, value := range request.Config {
				mergedValues[key] = value.Value
			}
		}

		if !reflect.DeepEqual(existingValues, mergedValues) {
			err := apiclient.SetConfig(self.ServiceId, mergedValues)
			if err != nil {
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
