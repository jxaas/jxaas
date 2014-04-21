package endpoints

import (
	"fmt"
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/core"
)

// The RPC endpoint is used by our own charms, and is not really RESTful (it is RPC)
type EndpointRpc struct {
}

// update_relation_properties RPC call
func (self *EndpointRpc) ItemUpdate_relation_properties() *EndpointRpcUpdateRelationProperties {
	child := &EndpointRpcUpdateRelationProperties{}
	return child
}

// update_relation_properties RPC call
type EndpointRpcUpdateRelationProperties struct {
}

// update_relation_properties RPC request
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

// Implement fmt.Stringer
func (self *RpcUpdateRelationPropertiesRequest) String() string {
	return log.AsJson(self)
}

// update_relation_properties RPC response
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

func parseService(s string) (tenant, serviceType, instanceId, module string, err error) {
	tokens := strings.SplitN(s, "-", 3)

	if len(tokens) != 3 {
		err = fmt.Errorf("Cannot parse service")
		return
	}

	if !strings.HasPrefix(tokens[0], "u") {
		err = fmt.Errorf("Cannot parse tenant")
		return
	}

	tenant = tokens[0][1:]
	serviceType = tokens[1]

	tail := tokens[2]
	lastDash := strings.LastIndex(tail, "-")
	if lastDash == -1 {
		instanceId = tail
		module = ""
	} else {
		instanceId = tail[:lastDash]
		module = tail[lastDash:]
	}

	return
}

func parseUnit(s string) (tenant, serviceType, instanceId, module, unitId string, err error) {
	lastSlash := strings.LastIndex(s, "/")

	var serviceSpec string
	if lastSlash != -1 {
		unitId = s[lastSlash+1:]
		serviceSpec = s[:lastSlash]
	} else {
		unitId = ""
		serviceSpec = s
	}

	tenant, serviceType, instanceId, module, err = parseService(serviceSpec)
	return
}

//func coalesce(p *string, alternative string) string {
//	if p == nil {
//		return alternative
//	}
//	return *p
//}

// update_relation_properties RPC handler
func (self *EndpointRpcUpdateRelationProperties) HttpPost(huddle *core.Huddle, request *RpcUpdateRelationPropertiesRequest) (*RpcUpdateRelationPropertiesResponse, error) {
	// TODO: Validate that this is coming from one of our machines?

	log.Info("Got RPC request: UpdateRelationProperties: %v", request)

	response := &RpcUpdateRelationPropertiesResponse{}

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
	tenant, serviceType, instanceId, _, unitId, err := parseUnit(remoteUnit)
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

	if request.Action == "broken" {
		err = instance.DeleteRelationInfo(unitId, relationId)
	} else {
		err = instance.SetRelationInfo(unitId, relationId, request.Properties)
	}

	//	err := apiclient.SetRelationInfo(primaryServiceName, unitId, relationId, request.Properties)

	if err != nil {
		return nil, err
	}

	return response, nil
}
