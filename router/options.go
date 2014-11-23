package router

import (
	"flag"
	"net/url"

	"github.com/justinsb/gova/log"
)

var (
	flagRegistryUrl = flag.String("registry", "etcd://127.0.0.1", "Registry location")
	flagListen      = flag.String("listen", ":8080", "listen on address")
)

type Options struct {
	Registry       RouterRegistry
	Listen         string
}

func GetOptions() *Options {
	flag.Parse()

	self := &Options{}

	self.Listen = *flagListen

	registryUrl, err := url.Parse(*flagRegistryUrl)
	if err != nil {
		log.Warn("Unable to parse registry url: %v", *flagRegistryUrl)
		return nil
	}
	if registryUrl.Host == "etcd" {
		registry, err := NewEtcdRouterRegistry(registryUrl)
		if err != nil {
			log.Warn("Unable to build etcd registry", err)
			return nil
		}
		self.Registry = registry
	} else {
		log.Warn("Unknown registry type: %v", registryUrl.Host)
		return nil
	}

	return self
}
