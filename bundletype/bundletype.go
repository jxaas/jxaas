package bundletype

import (
	"bytes"
	"strconv"
	"text/template"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/checks"
	"github.com/jxaas/jxaas/model"
)

type BundleType interface {
	Key() string

	PrimaryJujuService() string
	PrimaryRelationKey() string

	MapCfCredentials(bundle *bundle.Bundle, relationInfo *model.RelationInfo) (map[string]string, error)

	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
	IsStarted(annotations map[string]string) bool

	// Lets the bundle modify the relations that are returned
	BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error
	GetHealthChecks(bundle *bundle.Bundle) (map[string]jxaas.HealthCheck, error)

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
	key                string
	primaryRelationKey string
	bundleStore        *bundle.BundleStore
}

func (self *baseBundleType) Key() string {
	return self.key
}

func (self *baseBundleType) PrimaryJujuService() string {
	return self.key
}

func (self *baseBundleType) PrimaryRelationKey() string {
	return self.primaryRelationKey
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
				instanceValue := data.InstanceConfig.Options["protocol"]
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

func (self *baseBundleType) GetHealthChecks(bundle *bundle.Bundle) (map[string]jxaas.HealthCheck, error) {
	mapped := map[string]jxaas.HealthCheck{}
	for k, definition := range bundle.HealthChecks {
		if definition.Service != "" {
			check := &checks.ServiceHealthCheck{}
			check.ServiceName = definition.Service
			mapped[k] = check
		}
	}

	return mapped, nil
}

func runTemplate(templateDefinition string, context map[string]string) (string, error) {
	// TODO: Cache templates
	t, err := template.New(templateDefinition).Parse(templateDefinition)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, context)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (self *baseBundleType) MapCfCredentials(bundle *bundle.Bundle, relationInfo *model.RelationInfo) (map[string]string, error) {
	credentials := map[string]string{}

	template := map[string]string{}
	for k, v := range relationInfo.Properties {
		template[k] = v
	}

	for k, v := range bundle.CloudFoundryConfig.Credentials {
		substituted, err := runTemplate(v, template)
		if err != nil {
			log.Warn("Error while running template: %v=%v", k, v, err)
			return nil, err
		}
		credentials[k] = substituted
	}

	return credentials, nil
}

func (self *baseBundleType) GetDefaultScalingPolicy() *model.ScalingPolicy {
	policy := &model.ScalingPolicy{}
	return policy
}
