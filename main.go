package main

import (
	"math/rand"
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
		if service.PublicAddress == "" {
			log.Info("Service not ready (no public address): %v", key)
			return false
		}
	}
	return true
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

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
		huddle, err := core.NewHuddle(system, bundleStore, apiclient, privateUrl)
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
