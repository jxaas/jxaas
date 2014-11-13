package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewMysqlBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "mysql")
}
