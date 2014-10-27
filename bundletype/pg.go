package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
)

type PgBundleType struct {
	baseBundleType
}

func NewPgBundleType(bundleStore *bundle.BundleStore) *PgBundleType {
	self := &PgBundleType{}
	self.key = "pg"
	self.primaryRelationKey = "pgsql"
	self.bundleStore = bundleStore
	return self
}

func (self *PgBundleType) IsStarted(allAnnotations map[string]map[string]string) bool {
	// TODO: Loop over all when no primaryRelationKey?
	annotations := allAnnotations[self.primaryRelationKey]

	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__database") {
			annotationsReady = true
		}
	}

	return annotationsReady
}
