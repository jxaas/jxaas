package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type ElasticsearchBundleType struct {
	baseBundleType
}

func NewElasticsearchBundleType(bundleStore *bundle.BundleStore) *ElasticsearchBundleType {
	self := &ElasticsearchBundleType{}
	self.key = "es"
	self.bundleStore = bundleStore
	return self
}

func (self *ElasticsearchBundleType) IsStarted(annotations map[string]string) bool {
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__cluster-name") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *ElasticsearchBundleType) BuildRelationInfo(relationInfo *model.RelationInfo, data *RelationBuilder) {
	data.Relation = "elasticsearch"

	self.baseBundleType.BuildRelationInfo(relationInfo, data)
}
