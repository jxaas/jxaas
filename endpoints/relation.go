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

func coalesce(p *string, alternative string) string {
	if p == nil {
		return alternative
	}
	return *p
}

func (self *EndpointRelation) HttpPost(apiclient *juju.Client, relationInfo *model.RelationInfo) (*model.RelationInfo, error) {
	// TODO: Validate that this is coming from one of our machines?

	// Sanitize
	if relationInfo.Properties == nil {
		relationInfo.Properties = make(map[string]string)
	}

	service := self.Service()

	serviceName := service.PrimaryServiceName()
	unitId := coalesce(relationInfo.UnitId, "")
	relationId := coalesce(relationInfo.RelationId, "")
	err := apiclient.SetRelationInfo(serviceName, unitId, relationId, relationInfo.Properties)

	if err != nil {
		return nil, err
	}

	return self.HttpGet(apiclient)
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
