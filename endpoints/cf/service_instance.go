package cf

import (
	"net/http"
	"time"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/model"
)

type EndpointServiceInstance struct {
	Parent *EndpointServiceInstances
	Id     string
}

func (self *EndpointServiceInstance) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

// XXX: This needs to map service_bindings... may need to use Item
func (self *EndpointServiceInstance) ItemServiceBindings() *EndpointServiceBindings {
	child := &EndpointServiceBindings{}
	child.Parent = self
	return child
}

func (self *EndpointServiceInstance) getInstanceId() string {
	return self.Id
}

func (self *EndpointServiceInstance) HttpPut(request *CfCreateInstanceRequest) (*rs.HttpResponse, error) {
	helper := self.getHelper()

	log.Info("CF instance put request: %v", request)

	instance := helper.getInstance(request.ServiceId, self.Id)
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	// XXX: Support multiple plans?
	configureRequest := &model.Instance{}

	err := instance.Configure(configureRequest)
	if err != nil {
		return nil, err
	}

	// XXX: Copied from instance.go HttpPut; refactor into Configure call?
	for i := 1; i <= 10; i++ {
		model, err := instance.GetState()

		if err == nil && model == nil {
			time.Sleep(time.Second * 1)
			continue
		} else {
			if err != nil {
				return nil, err
			}

			response := &CfCreateInstanceResponse{}
			// XXX: We need a dashboard URL - maybe a Juju GUI?
			response.DashboardUrl = "http://localhost:8080"

			httpResponse := &rs.HttpResponse{Status: http.StatusCreated}
			httpResponse.Content = response
			return httpResponse, nil
		}
	}

	log.Warn("Unable to retrieve instance state, even after retries")
	return nil, rs.ErrNotFound()
}

func (self *EndpointServiceInstance) HttpDelete(httpRequest *http.Request) (*CfDeleteInstanceResponse, error) {
	helper := self.getHelper()

	queryValues := httpRequest.URL.Query()
	serviceId := queryValues.Get("service_id")
	//	planId := queryValues.Get("plan_id")

	instance := helper.getInstance(serviceId, self.getInstanceId())
	if instance == nil {
		return nil, rs.ErrNotFound()
	}

	err := instance.Delete()
	if err != nil {
		return nil, err
	}

	// XXX: Wait for deletion?

	response := &CfDeleteInstanceResponse{}
	return response, nil
}

type CfCreateInstanceRequest struct {
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
	OrganizationGuid string `json:"organization_guid"`
	SpaceGuid        string `json:"space_guid"`
}

type CfCreateInstanceResponse struct {
	DashboardUrl string `json:"dashboard_url"`
}

type CfDeleteInstanceResponse struct {
}
