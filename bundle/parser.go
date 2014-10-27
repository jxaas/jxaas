package bundle

import (
	"fmt"
	"strconv"

	"launchpad.net/goyaml"

	"github.com/justinsb/gova/log"
)

func asString(v interface{}) string {
	if v == nil {
		return ""
	}

	return fmt.Sprint(v)
}

func getString(config map[interface{}]interface{}, key string) string {
	v, found := config[key]
	if !found {
		return ""
	}
	return asString(v)
}

func getInt(config map[interface{}]interface{}, key string, defaultValue int) (int, error) {
	v, found := config[key]
	if !found {
		return defaultValue, nil
	}

	s := asString(v)

	if s == "<<" {
		return IMPLICIT_MARKER_INT, nil
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func getBool(config map[interface{}]interface{}, key string, defaultValue bool) (bool, error) {
	v, found := config[key]
	if !found {
		return defaultValue, nil
	}

	s := asString(v)
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return b, nil
}

func getStringMap(config map[interface{}]interface{}, key string) map[string]string {
	v, found := config[key]
	if !found {
		return nil
	}

	vMap, ok := v.(map[interface{}]interface{})
	if !ok {
		log.Warn("Expected generic map, found %T", v)
		return nil
	}

	out := map[string]string{}
	for key, v := range vMap {
		out[asString(key)] = asString(v)
	}
	return out
}

func getStringArray(config map[interface{}]interface{}, key string) []string {
	v, found := config[key]
	if !found {
		return []string{}
	}

	vList, ok := v.([]interface{})
	if !ok {
		log.Warn("Expected generic array, found %T", v)
		return nil
	}

	out := []string{}
	for _, v := range vList {
		out = append(out, asString(v))
	}
	return out
}

func parseServiceConfig(config interface{}) (*ServiceConfig, error) {
	var err error

	configMap, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected generic map for service, found %T", config)
	}

	self := &ServiceConfig{}
	self.Charm = getString(configMap, "charm")
	self.Branch = getString(configMap, "branch")

	self.NumberUnits, err = getInt(configMap, "num_units", 1)
	if err != nil {
		return nil, err
	}

	self.Exposed, err = getBool(configMap, "exposed", false)
	if err != nil {
		return nil, err
	}

	//	self.OpenPorts = getStringArray(configMap, "open_ports")
	self.Options = getStringMap(configMap, "options")
	return self, nil
}

func parseRelationConfig(config interface{}) (*RelationConfig, error) {
	configList, ok := config.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected generic list for relation, found %T", config)
	}

	self := &RelationConfig{}
	if len(configList) != 2 {
		return nil, fmt.Errorf("Expected 2 items for relation, found: %v", configList)
	}

	self.From = asString(configList[0])
	self.To = asString(configList[1])
	return self, nil
}

func parseHealthCheck(config interface{}) (*HealthCheckConfig, error) {
	configMap, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected generic map for health check, found %T", config)
	}

	self := &HealthCheckConfig{}
	self.Service = getString(configMap, "service")
	return self, nil
}

func ParseBundle(yaml string) (map[string]*Bundle, error) {
	config := map[string]interface{}{}
	err := goyaml.Unmarshal([]byte(yaml), &config)
	if err != nil {
		return nil, err
	}

	bundles := map[string]*Bundle{}

	for key, v := range config {
		bundle, err := parseBundleSection(v)
		if err != nil {
			return nil, err
		}
		bundles[key] = bundle
	}

	return bundles, nil
}

func parseBundleSection(data interface{}) (*Bundle, error) {
	var err error

	dataMap, ok := data.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected generic map for top-level, found %T", data)
	}

	self := &Bundle{}
	self.Services = map[string]*ServiceConfig{}
	services := dataMap["services"]
	if services == nil {
		return nil, fmt.Errorf("Expected services section")
	}
	serviceMap, ok := services.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected generic map for services, found %T", services)
	}
	for serviceKey, serviceDefinition := range serviceMap {
		self.Services[asString(serviceKey)], err = parseServiceConfig(serviceDefinition)
		if err != nil {
			return nil, err
		}
	}

	self.Relations = []*RelationConfig{}
	relations := dataMap["relations"]
	if relations != nil {
		relationList, ok := relations.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Expected generic list for relations, found %T", relations)
		}
		for _, relationDefinition := range relationList {
			relation, err := parseRelationConfig(relationDefinition)
			if err != nil {
				return nil, err
			}
			self.Relations = append(self.Relations, relation)
		}
	}

	self.HealthChecks = map[string]*HealthCheckConfig{}
	healthChecks := dataMap["checks"]
	if healthChecks != nil {
		healthChecksMap, ok := healthChecks.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Expected generic map for health checks, found %T", services)
		}
		for healthCheckKey, healthCheckDefinition := range healthChecksMap {
			self.HealthChecks[asString(healthCheckKey)], err = parseHealthCheck(healthCheckDefinition)
			if err != nil {
				return nil, err
			}
		}
	}

	self.CloudFoundryConfig = &CloudFoundryConfig{}
	cfConfig := dataMap["cloudfoundry"]
	if cfConfig != nil {
		cfConfigMap, ok := cfConfig.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Expected generic map for cloudfoundry, found %T", cfConfig)
		}
		self.CloudFoundryConfig.Credentials = getStringMap(cfConfigMap, "credentials")
	}

	self.Properties = getStringMap(dataMap, "properties")

	return self, nil
}
