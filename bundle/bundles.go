package bundle

import (
	"fmt"
	"strconv"

	"github.com/justinsb/gova/log"
)

type PortAssigner interface {
	AssignPort() (int, error)
	GetAssignedPort() (int, bool)
}

type ProxySettings struct {
	Host string
	Port int
}

type TemplateContext struct {
	SystemServices  map[string]string
	SystemImplicits map[string]string

	// The configuration options as specified by the user
	Options map[string]string

	Relations map[string]map[string]string

	NumberUnits int

	PublicPortAssigner PortAssigner

	Proxy *ProxySettings
}

func (self *TemplateContext) AssignPublicPort() (int, error) {
	if self.PublicPortAssigner == nil {
		return 0, fmt.Errorf("PublicPortAssigner not set")
	}

	return self.PublicPortAssigner.AssignPort()
}

func (self *TemplateContext) GetSpecialProperty(relationType, key, value string) string {
	// Some special cases
	// host, private-address map to the proxy host
	if key == "host" || key == "private-address" {
		// Use proxy address
		if self.Proxy != nil {
			value = self.Proxy.Host
		}
	}
	if key == "port" {
		// Use proxy port
		if self.Proxy != nil {
			value = strconv.Itoa(self.Proxy.Port)
		}
	}
	if key == "protocol" {
		instanceValue := self.Options["protocol"]
		if instanceValue != "" {
			value = instanceValue
		}
	}

	return value
}

type StubPortAssigner struct {
}

func (self *StubPortAssigner) AssignPort() (int, error) {
	return 0, nil
}

func (self *StubPortAssigner) GetAssignedPort() (int, bool) {
	return 0, false
}

func (self *BundleTemplate) GetDefaultOptions() map[string]string {
	if self.options == nil {
		return nil
	}
	return self.options.Defaults
}

func (self *BundleTemplate) GetBundle(templateContext *TemplateContext, tenant, serviceType, name string) (*Bundle, error) {
	// Copy and apply the system prefix
	templateContextCopy := *templateContext

	systemServices := map[string]string{}
	if templateContextCopy.SystemServices != nil {
		for k, v := range templateContextCopy.SystemServices {
			systemServices[k] = SYSTEM_PREFIX + v
		}
	}
	templateContextCopy.SystemServices = systemServices

	if templateContextCopy.Options == nil {
		templateContextCopy.Options = map[string]string{}
	}

	config, err := self.executeTemplate(&templateContextCopy)
	if err != nil {
		return nil, err
	}

	bundle, err := parseBundle(config)
	if err != nil {
		return nil, err
	}

	for k, v := range self.GetDefaultOptions() {
		_, found := templateContext.Options[k]
		if !found {
			templateContext.Options[k] = v
		}
	}

	bundle.ApplyImplicits(&templateContextCopy)

	bundle.ApplyPrefix(tenant, serviceType, name)

	return bundle, nil
}

//func (self *BundleTemplate) GetRaw() (*Bundle, error) {
//	bundles, err := parseBundle(self.template.Raw())
//	if err != nil {
//		return nil, err
//	}
//
//	bundle, err := getOnly(bundles)
//	if err != nil {
//		return nil, err
//	}
//
//	return bundle, nil
//}

func (self *BundleStore) GetSystemBundle(key string) (*Bundle, error) {
	log.Debug("Getting system bundle %v", key)

	template, err := self.GetBundleTemplate(key)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, nil
	}

	context := &TemplateContext{}

	config, err := template.executeTemplate(context)
	if err != nil {
		return nil, err
	}

	bundle, err := parseBundle(config)
	if err != nil {
		return nil, err
	}

	//	bundle, err := getOnly(bundles)
	//	if err != nil {
	//		return nil, err
	//	}

	return bundle, nil
}

//func getOnly(bundles map[string]*Bundle) (*Bundle, error) {
//	if len(bundles) > 1 {
//		return nil, fmt.Errorf("Multiple sections not handled")
//	}
//
//	for _, v := range bundles {
//		return v, nil
//	}
//
//	return nil, nil
//}
