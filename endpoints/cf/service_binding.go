package cf

import (
	"net/http"

	"github.com/justinsb/gova/rs"
)

type EndpointServiceBinding struct {
	Parent *EndpointServiceBindings
	Id     string
}

func (self *EndpointServiceBinding) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointServiceBinding) getInstanceId() string {
	return self.Parent.getInstanceId()
}

func (self *EndpointServiceBinding) HttpPut(request *CfBindRequest) (*CfBindResponse, error) {
	helper := self.getHelper()

	instance := helper.getInstance(request.PlanId, self.getInstanceId())
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	relationInfo, err := instance.GetRelationInfo(relationKey)

	if err != nil {
		return nil, err
	}

	// XXX: Synchronous wait??

	if relationInfo == nil {
		return nil, rs.ErrNotFound()
	}

	response := &CfBindResponse{}
	response.Credentials = map[string]string{}

	// XXX: How to map?
	for k, v := range relationInfo.Properties {
		response.Credentials[k] = v
	}

	return response, nil
}

func (self *EndpointServiceBinding) HttpDelete(httpRequest *http.Request) (*CfUnbindResponse, error) {
	helper := self.getHelper()

	serviceId := httpRequest.getQueryParameter("service_id")
	planId := httpRequest.getQueryParameter("plan_id")

	instance := helper.getInstance(planId, self.getInstanceId())
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	// XXX: actually remove something?

	response := &CfUnbindResponse{}

	return response, nil
}

type CfBindRequest struct {
	ServiceId string
	PlanId    string
	AppGuid   string
}

type CfBindResponse struct {
	Credentials map[string]string
}

type CfUnbindResponse struct {
}
