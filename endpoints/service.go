package endpoints

import (
	"math/rand"
	"net/http"
	"strconv"

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

	log.Debug("Service state: %v", status)

	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil

	return MapToInstance(self.ServiceId, status), nil
}

func (self *EndpointService) HttpPut(apiclient *juju.Client) (*Instance, error) {
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

	serviceName := "service" + strconv.Itoa(rand.Int())

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
		serviceName,
		numUnits,
		configYAML,
	)

	if err != nil {
		return nil, err
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
