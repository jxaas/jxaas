package endpoints

import (
	"github.com/jxaas/jxaas/core"
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

func (self *EndpointRelation) getInstance() *core.Instance {
	return self.Parent.Parent.getInstance()
}

func (self *EndpointRelation) HttpGet(apiclient *juju.Client) (*model.RelationInfo, error) {
	instance := self.getInstance()

	relationKey := self.RelationKey
	return instance.GetRelationInfo(relationKey)
}
