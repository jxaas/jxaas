package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type PgBundleType struct {
	baseBundleType
}

func NewPgBundleType(bundleStore *bundle.BundleStore) *PgBundleType {
	self := &PgBundleType{}
	self.key = "pg"
	self.bundleStore = bundleStore
	return self
}

func (self *PgBundleType) IsStarted(annotations map[string]string) bool {
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__database") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *PgBundleType) BuildRelationInfo(relationInfo *model.RelationInfo, data *RelationBuilder) {
	data.Relation = "pgsql"

	self.baseBundleType.BuildRelationInfo(relationInfo, data)
}
