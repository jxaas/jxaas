package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
)

type CassandraBundleType struct {
	baseBundleType
}

func NewCassandraBundleType(bundleStore *bundle.BundleStore) *CassandraBundleType {
	self := &CassandraBundleType{}
	self.key = "cassandra"
	self.primaryRelationKey = "cassandra"
	self.bundleStore = bundleStore
	return self
}

func (self *CassandraBundleType)IsStarted(allAnnotations map[string]map[string]string) bool {
	// TODO: Loop over all when no primaryRelationKey?
	annotations := allAnnotations[self.primaryRelationKey]

	// TODO: This is a total hack... need to figure out when annotations are 'ready' and when not.
	// we probably should do this on set, either in the charms or in the SetAnnotations call
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "private-address") {
			annotationsReady = true
		}
	}

	return annotationsReady
}
