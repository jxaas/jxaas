package bundletype

import "github.com/jxaas/jxaas/bundle"

func NewMultitenantMysqlBundleType(bundleStore *bundle.BundleStore) (*GenericBundleType, error) {
	return buildGenericFromStore(bundleStore, "multimysql")
}
