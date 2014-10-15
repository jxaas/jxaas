package cf

import ()

type EndpointCatalog struct {
	Parent *EndpointCfRoot
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
		//		service.Description = bundle.Name + " service"
		service.Bindable = true
		//	Tags        []string
		//	Metadata    map[string]string
		//	Requires    []string

		service.Plans = []*CatalogModelPlan{}
		plan := &CatalogModelPlan{}
		plan.Id = service.Id + "::" + "default"
		plan.Name = "Default plan"
		//	Description     string
		//	Metadata        map[string]string
		//	Free            bool
		//	DashboardClient *CatalogModelDashboard
		service.Plans = append(service.Plans, plan)

		model.Services = append(model.Services, service)
	}

	return model, nil
}

type CatalogModel struct {
	Services []*CatalogModelService
}

type CatalogModelService struct {
	Id          string // guid
	Name        string // cli friendly
	Description string
	Bindable    bool
	Tags        []string
	Metadata    map[string]string
	Requires    []string
	Plans       []*CatalogModelPlan
}

type CatalogModelPlan struct {
	Id              string // guid
	Name            string // cli friendly
	Description     string
	Metadata        map[string]string
	Free            bool
	DashboardClient *CatalogModelDashboard
}

type CatalogModelDashboard struct {
	Id          string
	Secret      string
	RedirectUrl string
}
