package endpoints

type EndpointXaasPrivate struct {
}

func (self *EndpointXaasPrivate) ItemRpc() *EndpointRpc {
	child := &EndpointRpc{}
	return child
}
