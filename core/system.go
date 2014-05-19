package core

import (
	"github.com/jxaas/jxaas/bundletype"
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

func NewSystem() *System {
	self := &System{}
	self.BundleTypes = map[string]bundletype.BundleType{}
	self.Scheduler = scheduler.NewScheduler()
	return self
}
