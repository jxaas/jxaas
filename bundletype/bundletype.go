package bundletype

import (
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type BundleType interface {
	Key() string
	PrimaryJujuService() string
	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
	IsStarted(annotations map[string]string) bool

	BuildRelationInfo(relationInfo *model.RelationInfo, relation string, properties []model.RelationProperty)
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

func (self *baseBundleType) BuildRelationInfo(relationInfo *model.RelationInfo, relation string, properties []model.RelationProperty) *model.RelationInfo {
	if relation != "" {
		for _, property := range properties {
			if property.RelationType != relation {
				continue
			}

			relationInfo.Properties[property.Key] = property.Value
		}
	}

	return relationInfo
}
