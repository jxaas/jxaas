package bundletype

import "github.com/jxaas/jxaas/bundle"

type MultitenantMysqlBundleType struct {
	baseBundleType
}

func NewMultitenantMysqlBundleType(bundleStore *bundle.BundleStore) *MultitenantMysqlBundleType {
	self := &MultitenantMysqlBundleType{}
	self.key = "multimysql"
	self.primaryRelationKey = "mysql"
	self.bundleStore = bundleStore
	return self
}
