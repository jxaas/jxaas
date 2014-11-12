package bundletype

import "github.com/jxaas/jxaas/bundle"

type MysqlBundleType struct {
	baseBundleType
}

func NewMysqlBundleType(bundleStore *bundle.BundleStore) *MysqlBundleType {
	self := &MysqlBundleType{}
	self.key = "mysql"
	//	self.primaryRelationKey = "mysql"
	self.bundleStore = bundleStore
	return self
}
