package endpoints

import "strings"

type EndpointXaas struct {
}

func (self *EndpointXaas) Item(key string) *EndpointTenant {
	child := &EndpointTenant{}

	tenant := strings.Replace(key, "-", "", -1)
	child.Tenant = tenant

	return child
}
