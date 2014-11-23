package router

import (
	"encoding/json"
	"net/url"

	"github.com/coreos/go-etcd/etcd"

	"github.com/justinsb/gova/log"
)

type EtcdRouterRegistry struct {
	client *etcd.Client
	basePath string
}

type etcdRouterData struct {
	Tenant  string
	Backend string
}

func NewEtcdRouterRegistry(etcdUrl *url.URL) (*EtcdRouterRegistry, error) {
	self := &EtcdRouterRegistry{}

	hosts := []string{}
	hosts = append(hosts, etcdUrl.Host)

	self.client = etcd.NewClient(hosts)

	path := etcdUrl.Path
	_, err := self.client.CreateDir(path, 0)
	if err != nil {
		log.Warn("Error creating path in etcd: %v", path, err)
		return nil, err
	}

	self.basePath = etcdUrl.Path

	return self, nil
}

func (self*EtcdRouterRegistry) keyForTenant(tenant string) string {
	key := self.basePath
	key = key+tenant
	return key
}

func (self*EtcdRouterRegistry) GetBackendForTenant(tenant string) string {
	key := self.keyForTenant(tenant)

	response, err := self.client.Get(key, false, false)
	if err != nil {
		log.Warn("Error reading key from etcd: %v", key, err)
		return ""
	}

	node := response.Node
	if node == nil || node.Value == "" {
		log.Info("No contents for key from etcd: %v", key)
		return ""
	}

	decoded := &etcdRouterData{}
	err = json.Unmarshal([]byte(node.Value), decoded)
	if err != nil {
		log.Warn("Error parsing value from etcd: %v", node.Value, err)
		return ""
	}

	return decoded.Backend
}


func (self*EtcdRouterRegistry) SetBackendForTenant(tenant string, backend string) error {
	key := self.keyForTenant(tenant)

	data := &etcdRouterData{}
	data.Backend = backend
	data.Tenant = tenant
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
