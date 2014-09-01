package bundletype

import (
	"github.com/jxaas/jxaas"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type BundleType interface {
	Key() string
	PrimaryJujuService() string
	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
	IsStarted(annotations map[string]string) bool

	BuildRelationInfo(relationInfo *model.RelationInfo, data *RelationBuilder)
	GetHealthChecks() []jxaas.HealthCheck

	GetDefaultScalingPolicy() *model.ScalingPolicy
}

// RelationProperties passes the parameters for BuildRelationInfo
// Allows extensibility and avoids a huge parameter list
type RelationBuilder struct {
	Relation   string
	Properties []model.RelationProperty
	ProxyHost  string
	ProxyPort  int
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

func (self *baseBundleType) BuildRelationInfo(relationInfo *model.RelationInfo, data *RelationBuilder) {
	if data.Relation != "" {
		for _, property := range data.Properties {
			if property.RelationType != data.Relation {
				continue
			}

			relationInfo.Properties[property.Key] = property.Value
		}
	}
}

func (self *baseBundleType) GetHealthChecks() []jxaas.HealthCheck {
	return []jxaas.HealthCheck{}
}

func (self *baseBundleType) GetDefaultScalingPolicy() *model.ScalingPolicy {
	policy := &model.ScalingPolicy{}
	return policy
}
