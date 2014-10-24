package bundletype

import (
	"strconv"
	"strings"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/model"
)

type MultitenantMysqlBundleType struct {
	baseBundleType
}

func NewMultitenantMysqlBundleType(bundleStore *bundle.BundleStore) *MultitenantMysqlBundleType {
	self := &MultitenantMysqlBundleType{}
	self.key = "multimysql"
	self.primaryRelationKey = "mysql"
	self.bundleStore = bundleStore
	return self
}

func (self *MultitenantMysqlBundleType) IsStarted(annotations map[string]string) bool {
	annotationsReady := false
	for key, _ := range annotations {
		if strings.HasSuffix(key, "__password") {
			annotationsReady = true
		}
	}

	return annotationsReady
}

func (self *MultitenantMysqlBundleType) BuildRelationInfo(bundle *bundle.Bundle, relationInfo *model.RelationInfo, data *RelationBuilder) error {
		data.Relation = self.primaryRelationKey


	err := self.baseBundleType.BuildRelationInfo(bundle, relationInfo, data)
	if err != nil {
		return err
	}

	// To override the IP / port
	if data.ProxyHost != "" {
		proxyHost := data.ProxyHost
		relationInfo.Properties["host"] = proxyHost
		relationInfo.Properties["private-address"] = proxyHost
		relationInfo.Properties["port"] = strconv.Itoa(data.ProxyPort)
		relationInfo.PublicAddresses = []string{proxyHost}
	}

	return nil
}
