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
	CF_STATE_CREATING  = "creating"
	CF_STATE_AVAILABLE = "available"
)

type EndpointServiceInstance struct {
	Parent *EndpointServiceInstances
	Id     string
	Service string
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

	planId := request.PlanId

	if request.ServiceId != self.Service {
		return nil, rs.ErrNotFound()
	}
	
	bundleType, instance := helper.getInstance(self.Service, self.Id)
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
		cfPlanId := self.Service + "::" + cfPlan.Key
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
	// XXX: We need a dashboard URL - maybe a Juju GUI?
	response.DashboardUrl = "http://localhost:8080"
	response.State = CF_STATE_CREATING

	httpResponse := &rs.HttpResponse{Status: http.StatusCreated}
	httpResponse.Content = response
	return httpResponse, nil
}


func (self *EndpointServiceInstance) HttpGet() (*rs.HttpResponse, error) {
	helper := self.getHelper()

	log.Info("CF instance GET request: %v", self.Id)

	bundleType, instance := helper.getInstance(self.Service, self.Id)
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
	// XXX: We need a dashboard URL - maybe a Juju GUI?
	response.DashboardUrl = "http://localhost:8080"
	if ready {
		response.State = CF_STATE_AVAILABLE
	} else {
		response.State = CF_STATE_CREATING
	}
	httpResponse := &rs.HttpResponse{Status: http.StatusOK}
	httpResponse.Content = response
	return httpResponse, nil
}


func (self *EndpointServiceInstance) HttpDelete(httpRequest *http.Request) (*CfDeleteInstanceResponse, error) {
	helper := self.getHelper()

	queryValues := httpRequest.URL.Query()
	serviceId := queryValues.Get("service_id")
	//	planId := queryValues.Get("plan_id")
	
	if serviceId != self.Service {
		return nil, rs.ErrNotFound()
	}

	log.Info("Deleting item %v %v", serviceId, self.Id)

	bundletype, instance := helper.getInstance(serviceId, self.getInstanceId())
	if instance == nil || bundletype == nil {
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
	State        string `json:"state"`
}

type CfDeleteInstanceResponse struct {
}
