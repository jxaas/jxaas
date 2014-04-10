package bundle

import "strings"

// System units are shared; we don't apply the prefix to them
const (
	SYSTEM_PREFIX = "__system__"
)

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

func buildPrefix(tenant string, serviceType string, unit string) string {
	prefix := "u" + tenant + "-" + serviceType + "-" + unit + "-"
	return prefix
}

func (self *Bundle) ApplyPrefix(tenant string, serviceType string, unit string) {
	prefix := buildPrefix(tenant, serviceType, unit)

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
