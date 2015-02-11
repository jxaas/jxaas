package router

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/go-etcd/etcd"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/model"
)

var (
	etcdErrorKeyNotFound   = 100
	etcdErrorAlreadyExists = 105
)

type EtcdRouterRegistry struct {
	client   *etcd.Client
	basePath string
}

type etcdRouterData struct {
	Service string
	Tenant  string
	Backend string
}

func NewEtcdRouterRegistry(etcdUrl *url.URL) (*EtcdRouterRegistry, error) {
	self := &EtcdRouterRegistry{}

	hosts := []string{}
	hosts = append(hosts, "http://"+etcdUrl.Host+":4001")

	log.Debug("Using etcd hosts: %v", hosts)

	self.client = etcd.NewClient(hosts)

	path := etcdUrl.Path
	_, err := self.client.CreateDir(path, 0)
	if err != nil {
		etcdError, ok := err.(*etcd.EtcdError)
		if !ok || etcdError.ErrorCode != etcdErrorAlreadyExists {
			log.Warn("Error creating path in etcd: %v", path, err)
			return nil, err
		}
	}

	self.basePath = etcdUrl.Path

	return self, nil
}

func (self *EtcdRouterRegistry) keyForTenant(service string, tenant string) string {
	return self.basePath + "/tenant/" + tenant + "/" + service
}

func (self *EtcdRouterRegistry) keyForService(service string) string {
	return self.basePath + "/service/" + service
}

func (self *EtcdRouterRegistry) ListServicesForTenant(tenant string) (*model.Bundles, error) {
	children, err := self.listSubkeys(self.basePath + "/service/")
	if err != nil {
		log.Warn("Error listing subkeys in etcd", err)
		return nil, err
	}

	tenantChildren := []string{}
	if tenant != "" {
		tenantChildren, err = self.listSubkeys(self.basePath + "/tenant/" + tenant)
		if err != nil {
			log.Warn("Error listing subkeys in etcd", err)
			return nil, err
		}
	}

	bundles := &model.Bundles{}
	bundles.Bundles = []model.Bundle{}

	all := append(children, tenantChildren...)
	for _, child := range all {
		bundle := &model.Bundle{}
		bundle.Id = child
		bundle.Name = child
		bundles.Bundles = append(bundles.Bundles, *bundle)
	}
	return bundles, nil
}

func (self *EtcdRouterRegistry) GetBackendForTenant(service string, tenant *string) string {
	var data *etcdRouterData
	var err error

	if tenant != nil {
		key := self.keyForTenant(service, *tenant)
		data, err = self.read(key)
	}

	if data == nil && err == nil {
		key := self.keyForService(service)
		data, err = self.read(key)
	}

	if err != nil {
		log.Warn("Error reading from etcd", err)
		return ""
	}

	if data != nil {
		return data.Backend
	}

	return ""
}

func (self *EtcdRouterRegistry) SetBackendForTenant(service string, tenant string, backend string) error {
	key := self.keyForTenant(service, tenant)

	data := &etcdRouterData{}
	data.Backend = backend
	data.Service = service
	data.Tenant = tenant

	return self.put(key, data)
}

func (self *EtcdRouterRegistry) ListServices() ([]string, error) {
	children, err := self.listSubkeys(self.basePath + "/service/")
	if err != nil {
		log.Warn("Error listing subkeys in etcd", err)
		return nil, err
	}
	return children, nil
}

func (self *EtcdRouterRegistry) SetBackendForService(service string, backend string) error {
	key := self.keyForService(service)

	data := &etcdRouterData{}
	data.Backend = backend
	data.Service = service

	return self.put(key, data)
}

func (self *EtcdRouterRegistry) put(key string, data *etcdRouterData) error {
	json, err := json.Marshal(data)
	if err != nil {
		log.Warn("Error encoding value to json", err)
		return err
	}

	_, err = self.client.Set(key, string(json), 0)
	if err != nil {
		log.Warn("Error writing key to etcd: %v", key, err)
		return err
	}

	return nil
}

func (self *EtcdRouterRegistry) read(key string) (*etcdRouterData, error) {
	response, err := self.client.Get(key, false, false)
	if err != nil {
		etcdError, ok := err.(*etcd.EtcdError)
		if ok && etcdError.ErrorCode == etcdErrorKeyNotFound {
			log.Debug("Etcd key not found: %v", key)
			return nil, nil
		}

		log.Warn("Error reading key from etcd: %v", key, err)
		return nil, err
	}

	node := response.Node
	if node == nil || node.Value == "" {
		log.Info("No contents for key from etcd: %v", key)
		return nil, nil
	}

	decoded := &etcdRouterData{}
	err = json.Unmarshal([]byte(node.Value), decoded)
	if err != nil {
		log.Warn("Error parsing value from etcd: %v", node.Value, err)
		return nil, err
	}

	return decoded, nil
}

func (self *EtcdRouterRegistry) listSubkeys(key string) ([]string, error) {
	response, err := self.client.Get(key, false, false)
	if err != nil {
		etcdError, ok := err.(*etcd.EtcdError)
		if ok && etcdError.ErrorCode == etcdErrorKeyNotFound {
			log.Debug("Etcd key not found: %v", key)
			return []string{}, nil
		}

		log.Warn("Error reading key from etcd: %v", key, err)
		return nil, err
	}

	if response == nil || response.Node == nil || response.Node.Nodes == nil {
		log.Info("No children for key from etcd: %v", key)
		return []string{}, nil
	}

	names := []string{}
	for _, node := range response.Node.Nodes {
		nodeKey := node.Key
		if !strings.HasPrefix(nodeKey, key) {
			return nil, fmt.Errorf("Key without expected prefix: %v vs %v", nodeKey, key)
		}
		suffix := nodeKey[len(key):]
		names = append(names, suffix)
	}
	return names, nil
}
