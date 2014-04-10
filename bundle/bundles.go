package bundle

import (
	"bytes"
	"fmt"

	"github.com/justinsb/gova/log"
)

type TemplateContext struct {
	SystemServices map[string]string
}

func (self *BundleStore) GetBundle(templateContext *TemplateContext, tenant, serviceType, name string) (*Bundle, error) {
	// Copy and apply the system prefix
	templateContextCopy := *templateContext

	systemServices := map[string]string{}
	for k, v := range templateContextCopy.SystemServices {
		systemServices[k] = SYSTEM_PREFIX + v
	}
	templateContextCopy.SystemServices = systemServices

	template, err := self.getBundleTemplate(serviceType)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = template.Execute(&buffer, templateContextCopy)
	if err != nil {
		return nil, err
	}

	yaml := buffer.String()
	log.Debug("Bundle is:\n%v", yaml)

	bundles, err := ParseBundle(yaml)
	if err != nil {
		return nil, err
	}

	if len(bundles) > 1 {
		return nil, fmt.Errorf("Multiple sections not handled")
	}

	for _, v := range bundles {
		v.ApplyPrefix(tenant, serviceType, name)
		return v, nil
	}

	return nil, nil
}
