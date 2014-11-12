package bundletype

import "github.com/jxaas/jxaas/bundle"

type MongodbBundleType struct {
	baseBundleType
}

func NewMongodbBundleType(bundleStore *bundle.BundleStore) *MongodbBundleType {
	self := &MongodbBundleType{}
	self.key = "mongodb"
	//	self.primaryRelationKey = "mongodb"
	self.bundleStore = bundleStore
	//	self.readyProperty = "replset"
	return self
}
