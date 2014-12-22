package bundletype

import (
	"bytes"
	"text/template"

	"github.com/justinsb/gova/assert"
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

	MapCloudFoundryCredentials(relationInfo *model.RelationInfo) (map[string]string, error)
	GetCloudFoundryPlans() ([]*bundle.CloudFoundryPlan, error)

	GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error)
	IsStarted(annotations map[string]map[string]string) bool

	BuildRelationInfo(templateContext *bundle.TemplateContext, bundle *bundle.Bundle, relationKey string) (*model.RelationInfo, error)

	GetHealthChecks(bundle *bundle.Bundle) (map[string]jxaas.HealthCheck, error)

	GetDefaultScalingPolicy() *model.ScalingPolicy
}

type baseBundleType struct {
	key            string
	bundleTemplate *bundle.BundleTemplate

	meta bundle.BundleMeta
}

func (self *baseBundleType) Init() error {
	assert.That(self.bundleTemplate != nil)

	meta := self.bundleTemplate.GetMeta()
	if meta != nil {
		self.meta = *meta
	}

	if self.meta.ReadyProperty == "" {
		self.meta.ReadyProperty = "password"
	}

	if self.meta.PrimaryRelationKey == "" {
		self.meta.PrimaryRelationKey = self.key
	}

	return nil
}

func (self *baseBundleType) Key() string {
	return self.key
}

func (self *baseBundleType) PrimaryJujuService() string {
	return self.key
}

func (self *baseBundleType) PrimaryRelationKey() string {
	return self.meta.PrimaryRelationKey
}

func (self *baseBundleType) GetBundle(templateContext *bundle.TemplateContext, tenant, name string) (*bundle.Bundle, error) {
	bundleKey := self.Key()
	return self.bundleTemplate.GetBundle(templateContext, tenant, bundleKey, name)
}

func (self *baseBundleType) BuildRelationInfo(templateContext *bundle.TemplateContext, bundle *bundle.Bundle, relationKey string) (*model.RelationInfo, error) {
	log.Info("BuildRelationInfo with %v", templateContext)

	// Find the properties the juju charm is exposing
	relationProperties := templateContext.Relations[relationKey]

	// Map those properties using the definition
	provideProperties := map[string]string{}

	if len(bundle.Provides) == 0 {
		// No explicit provides => derive automatically
		for k, v := range relationProperties {
			v = templateContext.GetSpecialProperty(relationKey, k, v)
			provideProperties[k] = v
		}

		// Auto-populate required properties that we generate
		required := []string{"protocol", "port"}
		for _, k := range required {
			v, found := relationProperties[k]
			if !found {
				v = templateContext.GetSpecialProperty(relationKey, k, v)
			}
			provideProperties[k] = v
		}

	} else {
		definition, found := bundle.Provides[relationKey]
		if !found {
			// Explicit provides, but no definition => no relation
			log.Debug("Request for relation, but no definition found: %v", relationKey)
			return nil, nil
		}

		for k, v := range definition.Properties {
			provideProperties[k] = v
		}
	}

	relationInfo := &model.RelationInfo{}
	if templateContext.Proxy != nil {
		relationInfo.PublicAddresses = []string{templateContext.Proxy.Host}
	}
	relationInfo.Properties = provideProperties

	return relationInfo, nil
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

func (self *baseBundleType) MapCloudFoundryCredentials(relationInfo *model.RelationInfo) (map[string]string, error) {
	credentials := map[string]string{}

	template := map[string]string{}
	for k, v := range relationInfo.Properties {
		template[k] = v
	}

	// Note that we use the raw (not-expanded version)
	// This is to avoid a circular dependency
	// (the cloudfoundry section is treated specially)
	credentials, err := self.bundleTemplate.GetCloudFoundryCredentials(template)
	if err != nil {
		log.Warn("Error building cloudfoundry credentials", err)
		return nil, err
	}

	return credentials, nil
}

func (self *baseBundleType) GetCloudFoundryPlans() ([]*bundle.CloudFoundryPlan, error) {
	assert.That(self.bundleTemplate != nil)
	plans := self.bundleTemplate.GetCloudFoundryPlans()
	if plans == nil {
		plan := &bundle.CloudFoundryPlan{}
		plan.Key = "default"
		plans := []*bundle.CloudFoundryPlan{ plan }
		return plans, nil
	}
	return plans, nil
}

func (self *baseBundleType) GetDefaultScalingPolicy() *model.ScalingPolicy {
	policy := &model.ScalingPolicy{}
	return policy
}

func (self *baseBundleType) IsStarted(allAnnotations map[string]map[string]string) bool {
	// TODO: Loop over all when no primaryRelationKey?
	annotations := allAnnotations[self.PrimaryRelationKey()]

	readyProperty := self.meta.ReadyProperty

	// TODO: This is a total hack... need to figure out when annotations are 'ready' and when not.
	// we probably should do this on set, either in the charms or in the SetAnnotations call
	annotationsReady := false
	for key, _ := range annotations {
		if key == readyProperty {
			annotationsReady = true
		}
	}

	return annotationsReady
}
