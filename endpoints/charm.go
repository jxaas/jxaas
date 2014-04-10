package endpoints

import (
	"net/http"

	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/rs"
)

type EndpointCharm struct {
	Parent      *EndpointServices
	ServiceType string
}

func (self *EndpointCharm) Item(key string) *EndpointService {
	child := &EndpointService{}
	child.Parent = self
	child.ServiceKey = key
	return child
}

func (self *EndpointCharm) HttpGet(apiclient *juju.Client) ([]*model.Instance, error) {
	status, err := apiclient.ListServices()
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, rs.HttpError(http.StatusNotFound)
	}

	instances := make([]*model.Instance, 0)
	for key, state := range status.Services {
		//fmt.Printf("%v => %v\n\n", key, state)
		instance := model.MapToInstance(key, &state, nil)

		instances = append(instances, instance)
	}

	//fmt.Printf("%v", status)

	return instances, nil
}
