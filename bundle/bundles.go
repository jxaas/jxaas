package bundle

import (
	"bytes"
	"fmt"

	"github.com/justinsb/gova/log"
)

type PortAssigner interface {
	AssignPort() (int, error)
	GetAssignedPort() (int, bool)
}

type TemplateContext struct {
	SystemServices  map[string]string
	SystemImplicits map[string]string

	// The configuration options as specified by the user
	Options map[string]string

	NumberUnits int

	PublicPortAssigner PortAssigner
}

func (self *TemplateContext) AssignPublicPort() (int, error) {
	if self.PublicPortAssigner == nil {
		return 0, fmt.Errorf("PublicPortAssigner not set")
	}

	return self.PublicPortAssigner.AssignPort()
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
