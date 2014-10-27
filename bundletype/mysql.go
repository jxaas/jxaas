package bundletype

import (
	"strings"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type MysqlBundleType struct {
	baseBundleType
}

func NewMysqlBundleType(bundleStore *bundle.BundleStore) *MysqlBundleType {
	self := &MysqlBundleType{}
	self.key = "mysql"
	self.primaryRelationKey = "mysql"
	self.bundleStore = bundleStore
	return self
}

func (self *MysqlBundleType) IsStarted(annotations map[string]string) bool {
	// First call
	//__jxaas_relinfo_0_db:46__private-address:10.0.3.176 __jxaas_relinfo_0_db:46_timestamp:50]

	// Ready call
	//__jxaas_relinfo_0_db:46__database:u2c1f1c9f92d7481a8015fd6b53fb2f26-mysql-jk-proxy-proxyclient __jxaas_relinfo_0_db:46__host:10.0.3.176 __jxaas_relinfo_0_db:46__password:oozahghaicongoo __jxaas_relinfo_0_db:46__private-address:10.0.3.176 __jxaas_relinfo_0_db:46__slave:False __jxaas_relinfo_0_db:46__user:cahshaimesaecae __jxaas_relinfo_0_db:46_timestamp:0]

	// TODO: This is a total hack... need to figure out when annotations are 'ready' and when not.
	// we probably should do this on set, either in the charms or in the SetAnnotations call
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__password") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *MysqlBundleType) BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error {
	data.Relation = self.primaryRelationKey

	return self.baseBundleType.BuildRelationInfo(bundle, relationInfo, data)
	//
	//	// To override the IP / port
	//	if data.ProxyHost != "" {
	//		proxyHost := data.ProxyHost
	//		relationInfo.Properties["host"] = proxyHost
	//		relationInfo.Properties["private-address"] = proxyHost
	//		relationInfo.Properties["port"] = strconv.Itoa(data.ProxyPort)
	//		relationInfo.PublicAddresses = []string{proxyHost}
	//	}
}
