package main

import (
	"github.com/justinsb/gova/log"

	"bitbucket.org/jsantabarbara/jxaas/endpoints"
	"bitbucket.org/jsantabarbara/jxaas/inject"
	"bitbucket.org/jsantabarbara/jxaas/juju"
	"bitbucket.org/jsantabarbara/jxaas/rs"
)

func main() {
	juju.Init()

	binder := inject.NewBinder()
	binder.AddProvider(juju.ClientFactory)

	injector := binder.CreateInjector()

	rest := rs.NewRestServer()
	rest.AddEndpoint("/xaas/", (*endpoints.EndpointXaas)(nil))
	rest.AddInjector(injector)

	log.Fatal("Error serving HTTP", rest.ListenAndServe())
}
