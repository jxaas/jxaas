package cf

import (
	"github.com/justinsb/gova/log"
)

type EndpointCatalog struct {
	Parent *EndpointCfV2
}

func (self *EndpointCatalog) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointCatalog) getService() *EndpointCfService {
	return self.Parent.getService()
}

func (self *EndpointCatalog) HttpGet() (*CatalogModel, error) {
	helper := self.getHelper()
	service := self.getService()
	
	huddle := helper.getHuddle()
	bundles := huddle.System.ListBundleTypes()

	model := &CatalogModel{}
	model.Services = []*CatalogModelService{}

	for _, bundle := range bundles.Bundles {
		serviceModel := &CatalogModelService{}

		if bundle.Id != service.BundleId {
			continue
		}

		serviceModel.Id = helper.mapBundleTypeIdToCfServiceId(bundle.Id)
		serviceModel.Name = bundle.Name
		serviceModel.Description = bundle.Name + " service"
		serviceModel.Bindable = true
		serviceModel.Tags = []string{}
		//	Metadata    map[string]string
		serviceModel.Requires = []string{}

		bundleType := huddle.System.GetBundleType(bundle.Id)

		cfPlans, err := bundleType.GetCloudFoundryPlans()
		if err != nil {
			log.Warn("Error retrieving CloudFoundry plans for bundle %v", bundle.Id, err)
			return nil, err
		}

		serviceModel.Plans = []*CatalogModelPlan{}
		for _, cfPlan := range cfPlans {
			plan := &CatalogModelPlan{}
			plan.Id = serviceModel.Id + "::" + cfPlan.Key
			plan.Name = cfPlan.Key
			plan.Description = cfPlan.Key + " plan"
			//	Metadata        map[string]string
			//	Free            bool
			//	DashboardClient *CatalogModelDashboard
			serviceModel.Plans = append(serviceModel.Plans, plan)
		}
		model.Services = append(model.Services, serviceModel)
	}

	return model, nil
}

type CatalogModel struct {
	Services []*CatalogModelService `json:"services"`
}

type CatalogModelService struct {
	Id          string              `json:"id"`   // guid
	Name        string              `json:"name"` // cli friendly
	Description string              `json:"description"`
	Bindable    bool                `json:"bindable"`
	Tags        []string            `json:"tags"`
	Metadata    map[string]string   `json:"metadata"`
	Requires    []string            `json:"requires"`
	Plans       []*CatalogModelPlan `json:"plans"`
}

type CatalogModelPlan struct {
	Id              string                 `json:"id"`   // guid
	Name            string                 `json:"name"` // cli friendly
	Description     string                 `json:"description"`
	Metadata        map[string]string      `json:"metadata"`
	Free            bool                   `json:"free"`
	DashboardClient *CatalogModelDashboard `json:"dashboard_client"`
}

type CatalogModelDashboard struct {
	Id          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectUri string `json:"redirect_uri"`
}
