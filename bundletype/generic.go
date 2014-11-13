package bundletype

import "github.com/jxaas/jxaas/bundle"

type GenericBundleType struct {
	baseBundleType
}

func NewGenericBundleType(key string, bundleTemplate *bundle.BundleTemplate) (*GenericBundleType, error) {
	self := &GenericBundleType{}
	self.key = key
	self.bundleTemplate = bundleTemplate

	err := self.Init()
	if err != nil {
		return nil, err
	}

	return self, nil
}

func buildGenericFromStore(bundleStore *bundle.BundleStore, key string) (*GenericBundleType, error) {
	bundleTemplate, err := bundleStore.GetBundleTemplate(key)
	if err != nil {
		return nil, err
	}
	return NewGenericBundleType(key, bundleTemplate)
}
