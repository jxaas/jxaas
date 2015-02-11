package cf

type EndpointCfRoot struct {
	Helper *CfHelper `inject:"y"`
}

func (self *EndpointCfRoot) getHelper() *CfHelper {
	return self.Helper
}

// XXX: We should probably authenticate against CloudFoundry!

func (self *EndpointCfRoot) Item(service string) *EndpointCfService {
	child := &EndpointCfService{}
	child.Parent = self
	return child
}
