package endpoints

import (
	"fmt"
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/core"
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
	BundleType  string
	ServiceName string
	Relation    string
	RelationId  string
	UnitId      string
	RemoteName  string
	Action      string
	Properties  map[string]string
}

func (self *RpcUpdateRelationPropertiesRequest) String() string {
	return log.AsJson(self)
}

type RpcUpdateRelationPropertiesResponse struct {
}

//func unitToService(unit string) string {
//	service := unit
//	lastSlash := strings.LastIndex(service, "/")
//	if lastSlash != -1 {
//		service = service[:lastSlash]
//	}
//	return service
//}

func parseService(s string) (tenant, serviceType, instanceId string, err error) {
	tokens := strings.SplitN(s, "-", 3)

	if len(tokens) != 3 {
		return "", "", "", fmt.Errorf("Cannot parse service")
	}

	if !strings.HasPrefix(tokens[0], "u") {
		err = fmt.Errorf("Cannot parse tenant")
		return
	}

	tenant = tokens[0][1:]
	serviceType = tokens[1]
	instanceId = tokens[2]

	return
}

func parseUnit(s string) (tenant, serviceType, instanceId, unitId string, err error) {
	lastSlash := strings.LastIndex(s, "/")

	var serviceSpec string
	if lastSlash != -1 {
		unitId = s[lastSlash+1:]
		serviceSpec = s[:lastSlash]
	} else {
		unitId = ""
		serviceSpec = s
	}

	tenant, serviceType, instanceId, err = parseService(serviceSpec)
	return
}

//func coalesce(p *string, alternative string) string {
//	if p == nil {
//		return alternative
//	}
//	return *p
//}

func (self *EndpointRpcUpdateRelationProperties) HttpPost(huddle *core.Huddle, request *RpcUpdateRelationPropertiesRequest) (*RpcUpdateRelationPropertiesResponse, error) {
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
	//	primaryServiceName := unitToService(remoteUnit)
	tenant, serviceType, instanceId, unitId, err := parseUnit(remoteUnit)
	if err != nil {
		return nil, err
	}

	//	service := self.Service()

	//	serviceName := service.PrimaryServiceName()
	//	unitId := coalesce(request.UnitId, "")
	//	relationId := coalesce(request.RelationId, "")

	instance := huddle.GetInstance(tenant, serviceType, instanceId)

	//	unitId := request.UnitId
	relationId := request.RelationId

	err = instance.SetRelationInfo(unitId, relationId, request.Properties)

	//	err := apiclient.SetRelationInfo(primaryServiceName, unitId, relationId, request.Properties)

	if err != nil {
		return nil, err
	}

	return response, nil
}
