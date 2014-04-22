package main

import (
	"net"
	"os"
	"time"

	"github.com/justinsb/gova/log"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/endpoints"
	"github.com/jxaas/jxaas/inject"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/rs"
)

func isHuddleReady(huddle *core.Huddle) bool {
	for key, service := range huddle.SharedServices {
		if service.PublicAddress == nil {
			log.Info("Service not ready: %v", key)
			return false
		}
	}
	return true
}

func buildHuddle(system *core.System, bundleStore *bundle.BundleStore, jujuApi *juju.Client, privateUrl string) (*core.Huddle, error) {
	key := "shared"

	systemBundle, err := bundleStore.GetSystemBundle(key)
	if err != nil {
		log.Warn("Error loading system bundle: %v", key, err)
		return nil, err
	}

	if systemBundle == nil {
		log.Warn("Cannot load system bundle: %v", key, err)
		return nil, nil
	}

	info, err := systemBundle.Deploy(jujuApi)
	if err != nil {
		log.Warn("Error deploying system bundle", err)
		return nil, err
	}

	huddle := &core.Huddle{}
	huddle.PrivateUrl = privateUrl
	huddle.SharedServices = map[string]*core.SharedService{}

	for key, service := range info.Services {
		sharedService := &core.SharedService{}
		sharedService.JujuName = key
		sharedService.Key = key

		status := service.Status
		if status != nil {
			for _, unit := range status.Units {
				if unit.PublicAddress != "" {
					sharedService.PublicAddress = net.ParseIP(unit.PublicAddress)
				}
			}
		}

		huddle.SharedServices[key] = sharedService
	}

	huddle.JujuClient = jujuApi
	huddle.System = system
	// TODO: Wait until initialized or offer a separate 'bootstrap' command

	return huddle, nil
}

func main() {
	juju.Init()

	binder := inject.NewBinder()
	binder.AddProvider(juju.ClientFactory)

	bundleStore := bundle.NewBundleStore("templates")
	binder.AddSingleton(bundleStore)

	apiclient, err := juju.ClientFactory()
	if err != nil {
		log.Fatal("Error building Juju client", err)
		os.Exit(1)
	}

	system := &core.System{}
	system.BundleTypes = map[string]bundletype.BundleType{}
	system.BundleTypes["mysql"] = bundletype.NewMysqlBundleType(bundleStore)
	system.BundleTypes["es"] = bundletype.NewElasticsearchBundleType(bundleStore)
	system.BundleTypes["mongodb"] = bundletype.NewMongodbBundleType(bundleStore)
	system.BundleTypes["pg"] = bundletype.NewPgBundleType(bundleStore)
	system.BundleTypes["multimysql"] = bundletype.NewMultitenantMysqlBundleType(bundleStore)

	// TODO: Use flag or autodetect
	privateUrl := "http://10.0.3.1:8080/xaasprivate"

	for {
		huddle, err := buildHuddle(system, bundleStore, apiclient, privateUrl)
		if err != nil {
			log.Fatal("Error building huddle", err)
			os.Exit(1)
		}
		if isHuddleReady(huddle) {
			log.Info("Huddle config is %v", huddle)
			binder.AddSingleton(huddle)
			break
		}
		time.Sleep(2 * time.Second)
	}

	injector := binder.CreateInjector()

	rest := rs.NewRestServer()
	rest.AddEndpoint("/xaas/", (*endpoints.EndpointXaas)(nil))
	rest.AddEndpoint("/xaasprivate/", (*endpoints.EndpointXaasPrivate)(nil))
	rest.WithInjector(injector)
	rest.AddReader(rs.NewJsonMessageBodyReader())
	rest.AddWriter(rs.NewJsonMessageBodyWriter())

	log.Info("Ready!")

	log.Fatal("Error serving HTTP", rest.ListenAndServe())
}
