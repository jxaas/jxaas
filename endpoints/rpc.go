package endpoints

import (
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/juju"
)

type EndpointRpc struct {
}

func (self *EndpointRpc) ItemUpdate_relation_properties() *EndpointRpcUpdateRelationProperties {
	child := &EndpointRpcUpdateRelationProperties{}
	return child
}

type EndpointRpcUpdateRelationProperties struct {
}

type RpcUpdateRelationPropertiesRequest struct {
	Tenant      string
	ServiceType string
	ServiceId   string
	Relation    string
	RelationId  string
	UnitId      string
	RemoteName  string
	Action      string
	Properties  map[string]string
}

type RpcUpdateRelationPropertiesResponse struct {
}

func unitToService(unit string) string {
	service := unit
	lastSlash := strings.LastIndex(service, "/")
	if lastSlash != -1 {
		service = service[:lastSlash]
	}
	return service
}

//func coalesce(p *string, alternative string) string {
//	if p == nil {
//		return alternative
//	}
//	return *p
//}

func (self *EndpointRpcUpdateRelationProperties) HttpPost(apiclient *juju.Client, request *RpcUpdateRelationPropertiesRequest) (*RpcUpdateRelationPropertiesResponse, error) {
	// TODO: Validate that this is coming from one of our machines?

	log.Info("Got RPC request: UpdateRelationProperties: %v", request)

	response := &RpcUpdateRelationPropertiesResponse{}

	if request.Action == "broken" {
		log.Info("Ignoring 'broken' action")
		return response, nil
	}

	// Sanitize
	if request.Properties == nil {
		request.Properties = make(map[string]string)
	}

	//	tenant := request.Tenant
	//	serviceType := request.ServiceType
	//	name := request.Name
	//	child := request.Child
	//
	//	primaryServiceName := buildQualifiedJujuName(tenant, serviceType, name, child)

	remoteUnit := request.RemoteName
	primaryServiceName := unitToService(remoteUnit)

	//	service := self.Service()

	//	serviceName := service.PrimaryServiceName()
	//	unitId := coalesce(request.UnitId, "")
	//	relationId := coalesce(request.RelationId, "")

	unitId := request.UnitId
	relationId := request.RelationId

	err := apiclient.SetRelationInfo(primaryServiceName, unitId, relationId, request.Properties)

	if err != nil {
		return nil, err
	}

	return response, nil
}
