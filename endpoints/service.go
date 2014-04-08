package endpoints

import (
	"net/http"
	"reflect"
	"strings"

	"launchpad.net/goyaml"

	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/rs"
	"github.com/justinsb/gova/log"
)

type EndpointService struct {
	Parent     *EndpointCharm
	ServiceKey string
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

func (self *EndpointService) ServiceName() string {
	tenant := self.Parent.Parent.Parent.Tenant
	tenant = strings.Replace(tenant, "-", "", -1)

	serviceType := self.Parent.ServiceType

	serviceKey := self.ServiceKey

	// The u prefix is for user.
	// This is both a way to separate out user services from our services,
	// and a way to make sure the service name is valid (is not purely numeric / does not start with a number)
	serviceName := "u" + tenant + "-" + serviceType + "-" + serviceKey
	return serviceName
}

func (self *EndpointService) HttpGet(apiclient *juju.Client) (*Instance, error) {
	serviceName := self.ServiceName()
	status, err := apiclient.GetStatus(serviceName)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	config, err := apiclient.FindConfig(serviceName)
	if err != nil {
		return nil, err
	}

	log.Debug("Service state: %v", status)

	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil

	return MapToInstance(serviceName, status, config), nil
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

func (self *EndpointService) HttpPut(apiclient *juju.Client, request *Instance) (*Instance, error) {
	serviceName := self.ServiceName()

	// Sanitize
	request.Id = ""
	request.Units = nil
	if request.Config == nil {
		request.Config = make(map[string]string)
	}
	request.ConfigParameters = nil

	config, err := apiclient.FindConfig(serviceName)
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

		configYaml, err := makeConfigYaml(serviceName, request.Config)
		if err != nil {
			return nil, err
		}

		log.Debug("Deploying with YAML: %v", configYaml)

		err = apiclient.ServiceDeploy(
			charmUrl,
			serviceName,
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
			err = apiclient.SetConfig(serviceName, mergedValues)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug("Configuration unchanged; won't reconfigure")
		}
	}

	if request.Exposed != nil {
		status, err := apiclient.GetStatus(serviceName)
		if err != nil {
			return nil, err
		}
		if status.Exposed != *request.Exposed {
			err = apiclient.SetExposed(serviceName, *request.Exposed)
			if err != nil {
				log.Warn("Error setting service to Exposed=%v", *request.Exposed, err)
				return nil, err
			}
		}
	}

	return self.HttpGet(apiclient)
}

func (self *EndpointService) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
	serviceName := self.ServiceName()

	err := apiclient.ServiceDestroy(serviceName)
	if err != nil {
		return nil, err
	}

	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
}
