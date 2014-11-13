package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewPgBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "pg")
}
