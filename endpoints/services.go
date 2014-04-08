package endpoints

import "bitbucket.org/jsantabarbara/jxaas/bundle"

type EndpointServices struct {
	Parent *EndpointTenant
}

func (self *EndpointServices) Item(key string) *EndpointCharm {
	child := &EndpointCharm{}
	child.Parent = self
	child.ServiceType = key
	return child
}

func (self *EndpointServices) HttpGet() (*bundle.Bundle, error) {

	// TODO: Use golang templating??

	//		parsed := bundle.ParseBundle(bundle)
	//		log.Info("Bundle: %v", parsed)
	//
	//		elasticsearch := "es1"
	//		bundle = strings.Replace(bundle, "{{ELASTICSEARCH}}", elasticsearch, -1)
	//
	//		prefix := "u" + self.Parent.Parent.Parent.Tenant
	//		bundle = strings.Replace(bundle, "{{PREFIX}}", prefix, -1)
	//
	//		key := self.ServiceKey
	//		bundle = strings.Replace(bundle, "{{KEY}}", key, -1)

	context := &bundle.TemplateContext{}
	context.SystemServices = map[string]string{}
	context.SystemServices["elasticsearch"] = "es1"

	tenant := self.Parent.Tenant
	service := "mysql"
	name := "mysql1"

	b, err := bundle.GetBundle("mysql", context, tenant, service, name)
	if err != nil {
		return nil, err
	}

	return b, nil
}
