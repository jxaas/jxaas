package bundletype

import "github.com/jxaas/jxaas/bundle"

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
