package bundletype

import (
	"strconv"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type BundleType interface {
	Key() string
	PrimaryJujuService() string
	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
	IsStarted(annotations map[string]string) bool

	// Lets the bundle modify the relations that are returned
	BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error
	GetHealthChecks() []jxaas.HealthCheck

	GetDefaultScalingPolicy() *model.ScalingPolicy
}

// RelationProperties passes the parameters for BuildRelationInfo
// Allows extensibility and avoids a huge parameter list
type RelationBuilder struct {
	Relation       string
	Properties     []model.RelationProperty
	ProxyHost      string
	ProxyPort      int
	InstanceConfig *model.Instance
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

func (self *baseBundleType) BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error {
	// TODO: Unclear if we should expose other properties... probably not
	if data.Relation != "" {
		for _, property := range data.Properties {
			if property.RelationType != data.Relation {
				continue
			}

			relationInfo.Properties[property.Key] = property.Value
		}
	}

	log.Info("BuildRelationInfo with %v", bundle.Properties)

	properties := bundle.Properties
	for k, v := range properties {
		propertyValue := relationInfo.Properties[k]

		if v == "<<" {
			if k == "host" || k == "private-address" {
				// Use proxy address
				if data.ProxyHost != "" {
					propertyValue = data.ProxyHost
				}
			}
			if k == "port" {
				// Use proxy port
				if data.ProxyHost != "" {
					propertyValue = strconv.Itoa(data.ProxyPort)
				}
			}
			if k == "protocol" {
				instanceValue := data.InstanceConfig.Config["protocol"]
				if instanceValue != "" {
					propertyValue = instanceValue
				}
			}
		} else {
			propertyValue = v
		}

		relationInfo.Properties[k] = propertyValue
	}

	if data.ProxyHost != "" {
		relationInfo.PublicAddresses = []string{data.ProxyHost}
	}

	return nil
}

func (self *baseBundleType) GetHealthChecks() []jxaas.HealthCheck {
	return []jxaas.HealthCheck{}
}

func (self *baseBundleType) GetDefaultScalingPolicy() *model.ScalingPolicy {
	policy := &model.ScalingPolicy{}
	return policy
}
