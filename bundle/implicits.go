package bundle

import "github.com/justinsb/gova/log"

const (
	IMPLICIT_MARKER     = "<<"
	IMPLICIT_MARKER_INT = -1
)

func (self *ServiceConfig) applyImplicits(templateContext *TemplateContext) {
	//	Options     map[string]string
	for k, v := range self.Options {
		if v == IMPLICIT_MARKER {
			option, found := templateContext.Options[k]
			if found {
				self.Options[k] = option
				continue
			}

			option, found = templateContext.SystemImplicits[k]
			if found {
				self.Options[k] = option
				continue
			}

			// Rely on the Juju default value
			delete(self.Options, k)
		}
	}

	//	NumberUnits int
	if self.NumberUnits == IMPLICIT_MARKER_INT {
		self.NumberUnits = templateContext.NumberUnits
	}

	//	Exposed     bool
}

func (self *RelationConfig) applyImplicits(templateContext *TemplateContext) {
}

func (self *ProvideConfig) applyImplicits(templateContext *TemplateContext, relationKey string) {
	relationInfo := templateContext.Relations[relationKey]

	//	Properties     map[string]string
	for k, v := range self.Properties {
		if v == IMPLICIT_MARKER {
			propertyValue := relationInfo[k]

			propertyValue = templateContext.GetSpecialProperty(relationKey, k, propertyValue)

			self.Properties[k] = propertyValue
		}
	}
}

func (self *Bundle) ApplyImplicits(templateContext *TemplateContext) {
	for _, v := range self.Services {
		v.applyImplicits(templateContext)
	}

	for _, v := range self.Relations {
		v.applyImplicits(templateContext)
	}

	for k, v := range self.Provides {
		v.applyImplicits(templateContext, k)
	}

	stub, found := self.Services["sc"]
	if found {
		self.configureStubClient(templateContext, stub)
		log.Info("Configured stubclient: %v", stub)
	} else {
		log.Info("sc (stubclient) not found")
	}
}

func (self *Bundle) configureStubClient(templateContext *TemplateContext, stub *ServiceConfig) {
	if stub.Options == nil {
		stub.Options = map[string]string{}
	}

	for _, key := range []string{"jxaas-privateurl", "jxaas-tenant", "jxaas-user", "jxaas-secret"} {
		stub.Options[key] = templateContext.SystemImplicits[key]
	}
}
