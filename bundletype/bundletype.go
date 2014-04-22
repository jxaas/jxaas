package bundletype

import "github.com/jxaas/jxaas/bundle"

type BundleType interface {
	Key() string
	PrimaryJujuService() string
	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
}

type baseBundleType struct {
	key         string
	bundleStore *bundle.BundleStore
}

func (self *baseBundleType) Key() string {
	return self.key
}

func (self *baseBundleType) PrimaryJujuService() string {
	return self.key
}

func (self *baseBundleType) GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error) {
	bundleKey := self.Key()
	return self.bundleStore.GetBundle(templateContext, tenant, bundleKey, name)
}
