package bundle

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/justinsb/gova/log"
)

type TemplateContext struct {
	SystemServices map[string]string
}

const (
	DEF_MYSQL = `sql: 
  services: 
    sql: 
      charm: "cs:~justin-fathomdb/precise/mysql-0"
      num_units: 1
    proxyclient: 
      charm: "cs:~justin-fathomdb/precise/proxy-client-0"
      num_units: 0
    metrics: 
      charm: "cs:~justin-fathomdb/precise/heka-collector-0"
      num_units: 0
  relations: 
    - - "proxyclient"
      - "mysql"
    - - "metrics"
      - "mysql"
    - - "metrics:elasticsearch"
      - "{{.SystemServices.elasticsearch}}:cluster"
`
)

func GetBundle(templateContext *TemplateContext, tenant, serviceType, name string) (*Bundle, error) {
	var def string

	// Copy and apply the system prefix
	templateContextCopy := *templateContext

	systemServices := map[string]string{}
	for k, v := range templateContextCopy.SystemServices {
		systemServices[k] = SYSTEM_PREFIX + v
	}
	templateContextCopy.SystemServices = systemServices

	// TODO: Load from file
	if serviceType == "mysql" {
		def = DEF_MYSQL
	}

	if def == "" {
		return nil, nil
	}

	// TODO: Cache templates
	tmpl, err := template.New("bundle-" + serviceType).Parse(def)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateContextCopy)
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
