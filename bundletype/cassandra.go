package bundletype

import "github.com/jxaas/jxaas/bundle"

// Simple example ... we could override methods in BundleType
func NewCassandraBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return LoadFromStore(bundleStore, "cassandra")
}
