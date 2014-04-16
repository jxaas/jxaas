package main

import (
	"net"
	"os"

	"github.com/justinsb/gova/log"

	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/endpoints"
	"github.com/jxaas/jxaas/inject"
	"github.com/jxaas/jxaas/juju"
	"github.com/jxaas/jxaas/rs"
)

func buildHuddle(bundleStore *bundle.BundleStore, jujuApi *juju.Client) (*core.Huddle, error) {
	key := "shared"

	systemBundle, err := bundleStore.GetSystemBundle(key)
	if err != nil {
		return nil, err
	}

	if systemBundle == nil {
		log.Warn("Cannot load system bundle: %v", key, err)
		return nil, nil
	}

	info, err := systemBundle.Deploy(jujuApi)
	if err != nil {
		return nil, err
	}

	huddle := &core.Huddle{}
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
	huddle, err := buildHuddle(bundleStore, apiclient)
	if err != nil {
		log.Fatal("Error building huddle", err)
		os.Exit(1)
	}
	binder.AddSingleton(huddle)

	log.Info("Huddle config is %v", huddle)

	injector := binder.CreateInjector()

	rest := rs.NewRestServer()
	rest.AddEndpoint("/xaas/", (*endpoints.EndpointXaas)(nil))
	rest.AddEndpoint("/xaasprivate/", (*endpoints.EndpointXaasPrivate)(nil))
	rest.WithInjector(injector)
	rest.AddReader(rs.NewJsonMessageBodyReader())
	rest.AddWriter(rs.NewJsonMessageBodyWriter())

	log.Fatal("Error serving HTTP", rest.ListenAndServe())
}
