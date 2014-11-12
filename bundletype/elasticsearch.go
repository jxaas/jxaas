package bundletype

import "github.com/jxaas/jxaas/bundle"

type ElasticsearchBundleType struct {
	baseBundleType
}

func NewElasticsearchBundleType(bundleStore *bundle.BundleStore) *ElasticsearchBundleType {
	self := &ElasticsearchBundleType{}
	self.key = "es"
//	self.primaryRelationKey = "elasticsearch"
	self.bundleStore = bundleStore
//	self.readyProperty = "cluster-name"
	return self
}
