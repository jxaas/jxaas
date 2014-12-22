package cf

import (
	"fmt"
	"net/http"

	"github.com/justinsb/gova/log"
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

func (self *EndpointServiceBinding) HttpPut(request *CfBindRequest) (*rs.HttpResponse, error) {
	helper := self.getHelper()

	instance := helper.getInstance(request.ServiceId, self.getInstanceId())
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	bundleType := helper.getBundleType(request.ServiceId)
	if bundleType == nil {
		log.Warn("BundleType not found for serviceId: %v", request.ServiceId)
		return nil, rs.ErrNotFound()
	}

	ready, err := waitReady(instance, 300)
	if err != nil {
		log.Warn("Error while waiting for instance to become ready", err)
		return nil, err
	}

	if !ready {
		log.Warn("Timeout waiting for service to be ready")
		return nil, fmt.Errorf("Service not ready")
	}

	relationKey := bundleType.PrimaryRelationKey()
	_, relationInfo, err := instance.GetRelationInfo(relationKey)
	if err != nil {
		return nil, err
	}

	if relationInfo == nil {
		return nil, rs.ErrNotFound()
	}

	credentials, err := bundleType.MapCfCredentials(relationInfo)
	if err != nil {
		log.Warn("Error mapping to CF", err)
		return nil, err
	}

	log.Debug("Relation info: %v", relationInfo)

	log.Debug("Mapped to CF credentials %v", credentials)

	response := &CfBindResponse{}
	response.Credentials = credentials

	httpResponse := &rs.HttpResponse{Status: http.StatusCreated}
	httpResponse.Content = response
	return httpResponse, nil
}

func (self *EndpointServiceBinding) HttpDelete(httpRequest *http.Request) (*CfUnbindResponse, error) {
	helper := self.getHelper()

	queryValues := httpRequest.URL.Query()
	serviceId := queryValues.Get("service_id")
	//	planId := queryValues.Get("plan_id")

	instance := helper.getInstance(serviceId, self.getInstanceId())
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	// XXX: actually remove something?

	response := &CfUnbindResponse{}

	return response, nil
}

type CfBindRequest struct {
	ServiceId string `json:"service_id"`
	PlanId    string `json:"plan_id"`
	AppGuid   string `json:"app_guid"`
}

type CfBindResponse struct {
	Credentials map[string]string `json:"credentials"`
}

type CfUnbindResponse struct {
}
