package bundletype

import (
	"strings"

	"github.com/jxaas/jxaas/bundle"
)

type MultitenantMysqlBundleType struct {
	baseBundleType
}

func NewMultitenantMysqlBundleType(bundleStore *bundle.BundleStore) *MultitenantMysqlBundleType {
	self := &MultitenantMysqlBundleType{}
	self.key = "multimysql"
	self.primaryRelationKey = "mysql"
	self.bundleStore = bundleStore
	return self
}

func (self *MultitenantMysqlBundleType) IsStarted(annotations map[string]string) bool {
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__password") {
			annotationsReady = true
		}
	}

	return annotationsReady
}
