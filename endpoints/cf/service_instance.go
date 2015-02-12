package cf

import (
	"net/http"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

const (
	CF_STATE_SUCCEEDED   = "succeeded"
	CF_STATE_FAILED      = "failed"
	CF_STATE_IN_PROGRESS = "in progress"
)

type EndpointServiceInstance struct {
	Parent *EndpointServiceInstances
	Id     string
}

func (self *EndpointServiceInstance) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointServiceInstance) getService() *EndpointCfService {
	return self.Parent.getService()
}

// TODO: This needs to map service_bindings... may need to use Item
func (self *EndpointServiceInstance) ItemServiceBindings() *EndpointServiceBindings {
	child := &EndpointServiceBindings{}
	child.Parent = self
	return child
}

func (self *EndpointServiceInstance) getInstanceId() string {
	return self.Id
}

func (self *EndpointServiceInstance) HttpPut(request *CfCreateInstanceRequest) (*rs.HttpResponse, error) {
	service := self.getService()

	log.Info("CF instance put request: %v", request)

	planId := request.PlanId
	cfServiceId := request.ServiceId

	if cfServiceId != service.CfServiceId {
		log.Warn("Service mismatch: %v vs %v", cfServiceId, service.CfServiceId)
		return nil, rs.ErrNotFound()
	}

	bundleType, instance := service.getInstance(self.Id)
	if instance == nil || bundleType == nil {
		return nil, rs.ErrNotFound()
	}

	cfPlans, err := bundleType.GetCloudFoundryPlans()
	if err != nil {
		log.Warn("Error retrieving CloudFoundry plans for bundle %v", bundleType, err)
		return nil, err
	}

	var foundPlan *bundle.CloudFoundryPlan
	for _, cfPlan := range cfPlans {
		cfPlanId := service.CfServiceId + "::" + cfPlan.Key
		if cfPlanId == planId {
			assert.That(foundPlan == nil)
			foundPlan = cfPlan
		}
	}

	if foundPlan == nil {
		log.Warn("Plan not found %v", planId)
		return nil, rs.ErrNotFound()
	}

	log.Debug("Found CF plan: %v", foundPlan)

	configureRequest := &model.Instance{}
	configureRequest.Options = foundPlan.Options

	err = instance.Configure(configureRequest)
	if err != nil {
		return nil, err
	}

	response := &CfCreateInstanceResponse{}
	// TODO: We need a dashboard URL - maybe a Juju GUI?
	response.DashboardUrl = "http://localhost:8080"
	response.State = CF_STATE_IN_PROGRESS
	response.LastOperation = &CfOperation{}
	response.LastOperation.State = CF_STATE_IN_PROGRESS

	log.Info("Sending response to CF service create", log.AsJson(response))

	httpResponse := &rs.HttpResponse{Status: http.StatusAccepted}
	httpResponse.Content = response
	return httpResponse, nil
}

func (self *EndpointServiceInstance) HttpGet() (*rs.HttpResponse, error) {
	service := self.getService()

	log.Info("CF instance GET request: %v", self.Id)

	bundleType, instance := service.getInstance(self.Id)
	if instance == nil || bundleType == nil {
		return nil, rs.ErrNotFound()
	}

	state, err := instance.GetState()
	if err != nil {
		log.Warn("Error while waiting for instance to become ready", err)
		return nil, err
	}

	ready := false
	if state == nil {
		log.Warn("Instance not yet created")
	} else {
		status := state.Status

		if status == "started" {
			ready = true
		} else if status == "pending" {
			ready = false
		} else {
			log.Warn("Unknown instance status: %v", status)
		}
	}

	response := &CfCreateInstanceResponse{}
	// TODO: We need a dashboard URL - maybe a Juju GUI?
	response.DashboardUrl = "http://localhost:8080"
	var cfState string
	if ready {
		cfState = CF_STATE_SUCCEEDED
	} else {
		cfState = CF_STATE_IN_PROGRESS
	}
	response.State = cfState
	response.LastOperation = &CfOperation{}
	response.LastOperation.State = cfState

	log.Info("Sending response to CF service get", log.AsJson(response))

	httpResponse := &rs.HttpResponse{Status: http.StatusOK}
	httpResponse.Content = response
	return httpResponse, nil
}

func (self *EndpointServiceInstance) HttpDelete(httpRequest *http.Request) (*CfDeleteInstanceResponse, error) {
	service := self.getService()

	queryValues := httpRequest.URL.Query()
	serviceId := queryValues.Get("service_id")
	//	planId := queryValues.Get("plan_id")

	if serviceId != service.CfServiceId {
		log.Warn("Service mismatch: %v vs %v", serviceId, service.CfServiceId)
		return nil, rs.ErrNotFound()
	}

	log.Info("Deleting item %v %v", serviceId, self.Id)

	bundletype, instance := service.getInstance(self.getInstanceId())
	if instance == nil || bundletype == nil {
		return nil, rs.ErrNotFound()
	}

	err := instance.Delete()
	if err != nil {
		return nil, err
	}

	// TODO: Wait for deletion?

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
	DashboardUrl  string       `json:"dashboard_url"`
	State         string       `json:"state"`
	LastOperation *CfOperation `json:"last_operation"`
}

type CfOperation struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

type CfDeleteInstanceResponse struct {
}
