package cf

type EndpointCatalog struct {
	Parent *EndpointCfV2
}

func (self *EndpointCatalog) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointCatalog) HttpGet() (*CatalogModel, error) {
	helper := self.getHelper()
	huddle := helper.getHuddle()
	bundles := huddle.System.ListBundleTypes()

	model := &CatalogModel{}
	model.Services = []*CatalogModelService{}

	for _, bundle := range bundles.Bundles {
		service := &CatalogModelService{}

		service.Id = helper.mapBundleTypeIdToCfServiceId(bundle.Id)
		service.Name = bundle.Name
		service.Description = bundle.Name + " service"
		service.Bindable = true
		service.Tags = []string{}
		//	Metadata    map[string]string
		service.Requires = []string{}

		service.Plans = []*CatalogModelPlan{}
		plan := &CatalogModelPlan{}
		plan.Id = service.Id + "::" + "default"
		plan.Name = "default"
		plan.Description = "Default plan"
		//	Metadata        map[string]string
		//	Free            bool
		//	DashboardClient *CatalogModelDashboard
		service.Plans = append(service.Plans, plan)

		model.Services = append(model.Services, service)
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
