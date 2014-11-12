package bundletype

import "github.com/jxaas/jxaas/bundle"

type CassandraBundleType struct {
	baseBundleType
}

func NewCassandraBundleType(bundleStore *bundle.BundleStore) *CassandraBundleType {
	self := &CassandraBundleType{}
	self.key = "cassandra"
//	self.primaryRelationKey = "cassandra"
	self.bundleStore = bundleStore
	//	self.readyProperty = "??private-address??"
	return self
}
