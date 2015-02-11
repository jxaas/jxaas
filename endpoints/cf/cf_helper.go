package cf

import (
	"strings"

	"github.com/justinsb/gova/log"

	"github.com/jxaas/jxaas/auth"
	"github.com/jxaas/jxaas/core"
)

type CfTenantIdMap struct {
	TenantId string
}

type CfHelper struct {
	Authenticator auth.Authenticator `inject:"y"`
	Huddle        *core.Huddle       `inject:"y"`
	TenantIdMap   *CfTenantIdMap     `inject:"y"`
}

func NewCfTenantIdMap(tenantId string) *CfTenantIdMap {
	self := &CfTenantIdMap{}
	self.TenantId = tenantId
	return self
}

func (self *CfHelper) getHuddle() *core.Huddle {
	return self.Huddle
}

func (self *CfHelper) getAuthenticator() auth.Authenticator {
	return self.Authenticator
}

func (self *CfHelper) mapBundleTypeIdToCfServiceId(bundleTypeId string) string {
	tenantId := self.TenantIdMap.TenantId
	prefix := tenantId + "::"

	return prefix + bundleTypeId
}

func (self *CfHelper) mapCfServiceIdToBundleTypeId(cfServiceId string) string {
	tenantId := self.TenantIdMap.TenantId
	prefix := tenantId + "::"

	if !strings.HasPrefix(cfServiceId, prefix) {
		log.Warn("CF serviceId not in our expected format: %v", cfServiceId)
		return ""
	}

	return cfServiceId[len(prefix):]

}
