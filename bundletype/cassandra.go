package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewCassandraBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "cassandra")
}
