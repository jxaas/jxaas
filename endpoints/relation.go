package endpoints

import (
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
)

type EndpointRelation struct {
	Parent      *EndpointRelations
	RelationKey string
}

func (self *EndpointRelation) getInstance() *core.Instance {
	return self.Parent.Parent.getInstance()
}

func (self *EndpointRelation) HttpGet(apiclient *juju.Client) (*model.RelationInfo, error) {
	instance := self.getInstance()

	relationKey := self.RelationKey
	_, relationInfo, err := instance.GetRelationInfo(relationKey)

	if err != nil {
		return nil, err
	}

	if relationInfo == nil {
		return nil, rs.ErrNotFound()
	}

	return relationInfo, err
}
