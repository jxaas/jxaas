package cf

import (
	"strings"

	"github.com/justinsb/gova/log"

	"github.com/jxaas/jxaas/auth"
	"github.com/jxaas/jxaas/core"
)

type CfHelper struct {
	Authenticator auth.Authenticator `inject:"y"`
	Huddle        *core.Huddle       `inject:"y"`
	TenantId      string
}

func (self *CfHelper) getHuddle() *core.Huddle {
	return self.Huddle
}

func (self *CfHelper) getAuthenticator() auth.Authenticator {
	return self.Authenticator
}

func (self *CfHelper) mapBundleTypeIdToCfServiceId(bundleTypeId string) string {
	prefix := self.Huddle.getUuid() + "::"

	return prefix + bundleTypeId
}

func (self *CfHelper) mapCfServiceIdToBundleTypeId(cfServiceId string) string {
	prefix := self.Huddle.getUuid() + "::"

	if !strings.HasPrefix(cfServiceId, prefix) {
		log.Warn("CF serviceId not in our expected format: %v", cfServiceId)
		return ""
	}

	return cfServiceId[len(prefix):]

}

func (self *CfHelper) getInstance(serviceId string, instanceId string) *core.Instance {
	huddle := self.Huddle

	bundleTypeId := self.mapCfServiceIdToBundleTypeId(serviceId)
	if bundleTypeId == "" {
		return nil
	}

	bundleType := huddle.System.GetBundleType(bundleTypeId)
	if bundleType == nil {
		log.Warn("Bundle type not found: %v", bundleTypeId)
		return nil
	}

	tenant := self.TenantId
	instance := huddle.NewInstance(tenant, bundleType, instanceId)
	return instance
}
