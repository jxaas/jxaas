package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
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

func (self *MongodbBundleType) IsStarted(annotations map[string]string) bool {
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__hostname") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *MongodbBundleType) BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error {
	data.Relation = self.primaryRelationKey

	return self.baseBundleType.BuildRelationInfo(bundle, relationInfo, data)
}
