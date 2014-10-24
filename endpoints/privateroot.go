package endpoints

type EndpointXaasPrivate struct {
}

// TODO: Add auth
// (and then add to python lib)
// (and then add to stubclient & maybe proxy)
func (self *EndpointXaasPrivate) ItemRpc() *EndpointRpc {
	child := &EndpointRpc{}
	return child
}
