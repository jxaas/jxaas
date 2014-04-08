package bundle

import (
	"fmt"
	"strconv"
	"strings"

	"launchpad.net/goyaml"

	"github.com/justinsb/gova/log"
)

// System units are shared; we don't apply the prefix to them
const (
	SYSTEM_PREFIX = "__system__"
)

type Bundle struct {
	Services  map[string]*ServiceConfig
	Relations []*RelationConfig
}

type ServiceConfig struct {
	Charm       string
	Branch      string
	NumberUnits int
	Options     map[string]string
}

type RelationConfig struct {
	From string
	To   string
}

func applyPrefix(key string, prefix string) string {
	if strings.HasPrefix(SYSTEM_PREFIX, key) {
		return key
	}
	return prefix + key
}

func (self *RelationConfig) applyPrefix(prefix string) {
	self.From = applyPrefix(self.From, prefix)
	self.To = applyPrefix(self.To, prefix)
}

func (self *Bundle) ApplyPrefix(tenant string, service string, unit string) {
	prefix := "u" + tenant + "-" + service + "-" + unit + "-"

	services := map[string]*ServiceConfig{}
	for k, v := range self.Services {
		k := applyPrefix(k, prefix)
		services[k] = v
	}
	self.Services = services

	for _, v := range self.Relations {
		v.applyPrefix(prefix)
	}
}

//wordpress-stage:
//    series: precise
//    services:
//        blog:
//            charm: wordpress
//            branch: lp:charms/precise/wordpress
//            constraints: mem=2
//            options:
//                tuning: optimized
//                engine: apache
//        db:
//            charm: mysql
//            branch: lp:charms/precise/mysql
//            options:
//                tuning-level: fast
//        memcached:
//            branch: lp:charms/precise/memcached
//            options:
//                request-limit: 32
//    relations:
//        - [blog, [db, memcached]]
//
//wordpress-prod:
//    series: precise
//    inherits: wordpress-stage
//    services:
//        blog:
//            options:
//                engine: nginx
//                tuning: optimized
//            constraints: cpu-cores=1
//
//        db:
//            constraints: cpu-cores=2
//            options:
//                tuning-level: safest

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
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
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

	return self, nil
}
