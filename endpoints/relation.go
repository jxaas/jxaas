package endpoints

import (
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type EndpointRelation struct {
	Parent      *EndpointRelations
	RelationKey string
}

func (self *EndpointRelation) Service() *EndpointService {
	return self.Parent.Parent
}

func (self *EndpointRelation) HttpGet(apiclient *juju.Client) (*model.RelationInfo, error) {
	service := self.Service()
	relationKey := self.RelationKey

	serviceName := service.PrimaryServiceName()
	relationInfo, err := apiclient.GetRelationInfo(serviceName, relationKey)
	if err != nil {
		return nil, err
	}

	return relationInfo, nil
}
