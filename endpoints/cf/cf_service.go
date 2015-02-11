package cf

import (
	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/core"
	"strings"
)

type EndpointCfService struct {
	Parent      *EndpointCfRoot
	CfServiceId string
	BundleId    string
}

func (self *EndpointCfService) getHelper() *CfHelper {
	return self.Parent.getHelper()
}

func (self *EndpointCfService) getService() *EndpointCfService {
	return self
}

func (self *EndpointCfService) ItemV2() *EndpointCfV2 {
	child := &EndpointCfV2{}
	child.Parent = self
	return child
}

func (self *EndpointCfService) getInstance(instanceId string) (bundletype.BundleType, *core.Instance) {
	helper := self.getHelper()
	huddle := helper.Huddle

	bundleType := self.getBundleType()
	if bundleType == nil {
		log.Warn("Bundle type not found: %v", self.BundleId)
		return nil, nil
	}

	instanceId = strings.Replace(instanceId, "-", "", -1)

	tenantId := helper.TenantIdMap.TenantId
	instance := huddle.NewInstance(tenantId, bundleType, instanceId)
	return bundleType, instance
}

func (self *EndpointCfService) getBundleType() bundletype.BundleType {
	helper := self.getHelper()
	huddle := helper.Huddle

	//	bundleTypeId := self.mapCfServiceIdToBundleTypeId(serviceId)
	//	if bundleTypeId == "" {
	//		return nil
	//	}

	bundleType := huddle.System.GetBundleType(self.BundleId)
	return bundleType
}
