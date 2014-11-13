package core

import (
	"fmt"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/sources"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/model"
	"github.com/jxaas/jxaas/scheduler"
)

// System is the top-level object for storing system state
type System struct {
	BundleTypes map[string]bundletype.BundleType
	Scheduler   *scheduler.Scheduler
}

// Gets the bundle type by key
func (self *System) GetBundleType(key string) bundletype.BundleType {
	return self.BundleTypes[key]
}

// Gets information on all registered bundle types
func (self *System) ListBundleTypes() *model.Bundles {
	bundles := &model.Bundles{}

	bundleList := []model.Bundle{}
	for key, _ := range self.BundleTypes {
		bundle := model.Bundle{}
		bundle.Id = key
		bundle.Name = key
		bundleList = append(bundleList, bundle)
	}

	bundles.Bundles = bundleList

	return bundles
}

// Adds a bundletype to the system
func (self *System) AddBundleType(bundleType bundletype.BundleType) {
	self.BundleTypes[bundleType.Key()] = bundleType
}

// Adds a bundletype to the system, by extracting the required template from the charm itself
func (self *System) AddJxaasCharm(apiclient *juju.Client, key string, charmName string) error {
	charmInfo, err := apiclient.CharmInfo(charmName)
	if err != nil {
		log.Warn("Error reading charm: %v", charmName, err)
		return err
	}

	if charmInfo == nil {
		return fmt.Errorf("Unable to find charm: %v", charmName)
	}

	url := charmInfo.URL
	if url == "" {
		return fmt.Errorf("Unable to find charm url: %v", charmName)
	}

	contents, err := apiclient.DownloadCharm(charmName)
	if err != nil {
		log.Warn("Error reading charm", err)
		return err
	}

	charmFile := NewCharmReader(contents)
	config, err := charmFile.read("jxaas.yaml")
	if err != nil {
		log.Warn("Error reading jxaas.yaml from charm: %v", charmName, err)
		return err
	}

	if config == nil {
		return fmt.Errorf("Could not find jxaas.yaml in charm: %v", charmName)
	}
	//	log.Info("Jxaas config: %v", string(config))

	bundleTemplate, err := bundle.NewBundleTemplate(sources.NewArrayToByteSource(config))
	if err != nil {
		return err
	}

	bundleType, err := bundletype.NewGenericBundleType(key, bundleTemplate)
	if err != nil {
		return err
	}

	self.BundleTypes[key] = bundleType

	return nil
}

func NewSystem() *System {
	self := &System{}
	self.BundleTypes = map[string]bundletype.BundleType{}
	self.Scheduler = scheduler.NewScheduler()
	return self
}
