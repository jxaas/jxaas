package bundletype

import "github.com/jxaas/jxaas/bundle"

type ElasticsearchBundleType struct {
}

func NewElasticsearchBundleType(bundleStore *bundle.BundleStore) *MysqlBundleType {
	self := &MysqlBundleType{}
	self.key = "es"
	self.bundleStore = bundleStore
	return self
}
