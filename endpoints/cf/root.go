package cf

type EndpointCfRoot struct {
	Helper *CfHelper `inject:"y"`
}

func (self *EndpointCfRoot) getHelper() *CfHelper {
	return self.Helper
}

// TODO: We should probably authenticate against CloudFoundry!

func (self *EndpointCfRoot) Item(bundleId string) *EndpointCfService {
	child := &EndpointCfService{}
	child.Parent = self

	child.BundleId = bundleId
	helper := self.getHelper()
	child.CfServiceId = helper.mapBundleTypeIdToCfServiceId(bundleId)

	return child
}
