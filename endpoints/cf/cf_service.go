package cf

type EndpointCfService struct {
	Parent *EndpointCfRoot
	Service string
}

func (self *EndpointCfService) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointCfService) ItemV2() *EndpointCfV2 {
	child := &EndpointCfV2{}
	child.Parent = self
	child.Service = self.Service
	return child
}
