package endpoints

import (
	"fmt"

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

	if remoteUnit == "" {
		// We're a bit stuck here.  We do have the relationId and other info,
		// we just don't have the remote relation, and we're storing the attributes on the remote relation
		// TODO: Infer the remote relation? (-stubclient to -primary)?
		log.Warn("No remote unit; can't remove relations")
		return response, nil
	}

	//	primaryServiceName := unitToService(remoteUnit)
	tenant, bundleTypeName, instanceId, _, unitId, err := core.ParseUnit(remoteUnit)
	if err != nil {
		return nil, err
	}

	//	service := self.Service()

	//	serviceName := service.PrimaryServiceName()
	//	unitId := coalesce(request.UnitId, "")
	//	relationId := coalesce(request.RelationId, "")

	bundleType := huddle.System.GetBundleType(bundleTypeName)
	if bundleType == nil {
		return nil, fmt.Errorf("Unknown bundle type: %v", bundleTypeName)
	}

	instance := huddle.GetInstance(tenant, bundleType, instanceId)

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
