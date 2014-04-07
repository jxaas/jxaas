package endpoints

import (
	"net/http"

	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/rs"
	"github.com/justinsb/gova/log"
)

type EndpointService struct {
	Parent    *EndpointCharm
	ServiceId string
}

//func (self *EndpointService) HttpGet() ([]*Instance, error) {
//	//	return "Hello world"
//	envName := cmd.ReadCurrentEnvironment()
//	apiclient, err := juju.NewAPIClientFromName(envName)
//	if err != nil {
//		return nil, fmt.Errorf(connectionError, envName, err)
//	}
//	defer apiclient.Close()
//
//	patterns := make([]string, 0)
//
//	status, err := apiclient.Status(patterns)
//
//	//	if params.IsCodeNotImplemented(err) {
//	//		logger.Infof("Status not supported by the API server, " +
//	//			"falling back to 1.16 compatibility mode " +
//	//			"(direct DB access)")
//	//		status, err = c.getStatus1dot16()
//	//	}
//	// Display any error, but continue to print status if some was returned
//	if err != nil {
//		return nil, err
//	}
//
//	instances := make([]*Instance, 0)
//	for key, state := range status.Services {
//		fmt.Printf("%v => %v\n\n", key, state)
//		instance := MapToInstance(key, &state)
//
//		instances = append(instances, instance)
//	}
//
//	fmt.Printf("%v", status)
//
//	return instances, nil
//	//
//	//	result := formatStatus(status)
//	//
//	//	return c.out.Write(ctx, result), nil
//}

//func (self *EndpointService) HttpPut() {
//
//}
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

	log.Debug("Service state: %v", status)

	return MapToInstance(self.ServiceId, status), nil
}

//func (self *EndpointService) HttpPost(apiclient *juju.Client) (*Instance, error) {
//	//	curl, err := charm.InferURL(c.CharmName, conf.DefaultSeries())
//	//	if err != nil {
//	//		return err
//	//	}
//	//	repo, err := charm.InferRepository(curl, ctx.AbsPath(c.RepoPath))
//	//	if err != nil {
//	//		return err
//	//	}
//	//
//	//	repo = config.SpecializeCharmRepo(repo, conf)
//	//
//	//	curl, err = addCharmViaAPI(client, ctx, curl, repo)
//	//	if err != nil {
//	//		return err
//	//	}
//
//	//
//	//	charmInfo, err := client.CharmInfo(curl.String())
//	//	if err != nil {
//	//		return err
//	//	}
//
//	numUnits := 1
//
//	serviceName := "service" + strconv.Itoa(rand.Int())
//
//	//	if serviceName == "" {
//	//		serviceName = charmInfo.Meta.Name
//	//	}
//
//	var configYAML []byte
//	//	if c.Config.Path != "" {
//	//		configYAML, err = c.Config.Read(ctx)
//	//		if err != nil {
//	//			return err
//	//		}
//	//	}
//	var constraints constraints.Value
//	var toMachineSpec string
//
//	charmUrl := "cs:precise/mysql-38"
//
//	err = apiclient.ServiceDeploy(
//		charmUrl,
//		serviceName,
//		numUnits,
//		string(configYAML),
//		constraints,
//		toMachineSpec,
//	)
//
//	//	if params.IsCodeNotImplemented(err) {
//	//		logger.Infof("Status not supported by the API server, " +
//	//			"falling back to 1.16 compatibility mode " +
//	//			"(direct DB access)")
//	//		status, err = c.getStatus1dot16()
//	//	}
//	// Display any error, but continue to print status if some was returned
//	if err != nil {
//		return nil, err
//	}
//
//	return self.Item(serviceName).HttpGet()
//}
//
//func (self *EndpointCharm) HttpDelete(apiclient *juju.Client) (*rs.HttpResponse, error) {
//	serviceName := self.Key
//
//	err = apiclient.ServiceDestroy(serviceName)
//	if err != nil {
//		return nil, err
//	}
//
//	return &rs.HttpResponse{Status: http.StatusAccepted}, nil
//}
