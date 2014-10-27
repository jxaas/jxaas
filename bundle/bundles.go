package bundle

import (
	"bytes"
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

func (self *BundleStore) GetBundle(templateContext *TemplateContext, tenant, serviceType, name string) (*Bundle, error) {
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

	template, err := self.getBundleTemplate(serviceType)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, nil
	}

	var buffer bytes.Buffer
	err = template.Execute(&buffer, &templateContextCopy)
	if err != nil {
		return nil, err
	}

	// TODO: Replace the "<no value>" incorrect placeholders
	// https://code.google.com/p/go/issues/detail?id=6288

	yaml := buffer.String()
	log.Debug("Bundle is:\n%v", yaml)

	bundles, err := ParseBundle(yaml)
	if err != nil {
		return nil, err
	}

	bundle, err := getOnly(bundles)
	if err != nil {
		return nil, err
	}

	bundle.ApplyImplicits(&templateContextCopy)

	bundle.ApplyPrefix(tenant, serviceType, name)

	return bundle, nil
}

func (self *BundleStore) GetSystemBundle(key string) (*Bundle, error) {
	template, err := self.getBundleTemplate(key)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, nil
	}

	context := make(map[string]string)

	var buffer bytes.Buffer
	err = template.Execute(&buffer, context)
	if err != nil {
		return nil, err
	}

	yaml := buffer.String()
	log.Debug("Bundle is:\n%v", yaml)

	bundles, err := ParseBundle(yaml)
	if err != nil {
		return nil, err
	}

	bundle, err := getOnly(bundles)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}

func getOnly(bundles map[string]*Bundle) (*Bundle, error) {
	if len(bundles) > 1 {
		return nil, fmt.Errorf("Multiple sections not handled")
	}

	for _, v := range bundles {
		return v, nil
	}

	return nil, nil
}
