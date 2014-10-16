package cf

type EndpointCfRoot struct {
	Helper *CfHelper `inject:"y"`
}

func (self *EndpointCfRoot) getHelper() *CfHelper {
	return self.Helper
}

// XXX: We should probably authenticate against CloudFoundry!

func (self *EndpointCfRoot) ItemV2() *EndpointCfV2 {
	child := &EndpointCfV2{}
	child.Parent = self
	return child
}
