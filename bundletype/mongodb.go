package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewMongodbBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "mongodb")
}
