package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
)

type ElasticsearchBundleType struct {
	baseBundleType
}

func NewElasticsearchBundleType(bundleStore *bundle.BundleStore) *ElasticsearchBundleType {
	self := &ElasticsearchBundleType{}
	self.key = "es"
	self.primaryRelationKey = "elasticsearch"
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
