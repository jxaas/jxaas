package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewElasticsearchBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "es")
}
