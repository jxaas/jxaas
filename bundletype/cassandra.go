package bundletype

import (
	"strconv"
	"strings"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type CassandraBundleType struct {
	baseBundleType
}

func NewCassandraBundleType(bundleStore *bundle.BundleStore) *CassandraBundleType {
	self := &CassandraBundleType{}
	self.key = "cassandra"
	self.primaryRelationKey = "cassandra"
	self.bundleStore = bundleStore
	return self
}

func (self *CassandraBundleType) IsStarted(annotations map[string]string) bool {
	// TODO: This is a total hack... need to figure out when annotations are 'ready' and when not.
	// we probably should do this on set, either in the charms or in the SetAnnotations call
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "private-address") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *CassandraBundleType) BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error {
	err := self.baseBundleType.BuildRelationInfo(bundle, relationInfo, data)
	if err != nil {
		return err
	}

	// To override the IP / port
	if data.ProxyHost != "" {
		proxyHost := data.ProxyHost
		relationInfo.Properties["host"] = proxyHost
		relationInfo.Properties["port"] = strconv.Itoa(data.ProxyPort)
		relationInfo.Properties["private-address"] = proxyHost
		relationInfo.PublicAddresses = []string{proxyHost}
	}

	return nil
}
