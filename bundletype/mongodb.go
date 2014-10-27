package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
)

type MongodbBundleType struct {
	baseBundleType
}

func NewMongodbBundleType(bundleStore *bundle.BundleStore) *MongodbBundleType {
	self := &MongodbBundleType{}
	self.key = "mongodb"
	self.primaryRelationKey = "mongodb"
	self.bundleStore = bundleStore
	return self
}

func (self *MongodbBundleType) IsStarted(allAnnotations map[string]map[string]string) bool {
	// TODO: Loop over all when no primaryRelationKey?
	annotations := allAnnotations[self.primaryRelationKey]

	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__hostname") {
			annotationsReady = true
		}
	}

	return annotationsReady
}
