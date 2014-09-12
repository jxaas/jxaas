package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"launchpad.net/goyaml"
	"launchpad.net/juju-core/state/api"

	"github.com/justinsb/gova/inject"
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"
	"github.com/jxaas/jxaas/bundle"
	"github.com/jxaas/jxaas/bundletype"
	"github.com/jxaas/jxaas/core"
	"github.com/jxaas/jxaas/endpoints"
	"github.com/jxaas/jxaas/juju"
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

	options := GetOptions()
	if options == nil {
		log.Fatal("Error reading options")
		os.Exit(1)
	}

	juju.Init()

	binder := inject.NewBinder()

	clientFactory := juju.EnvClientFactory

	if options.AgentConf != "" && options.ApiPasswordPath != "" {
		yaml, err := ioutil.ReadFile(options.AgentConf)
		if err != nil {
			log.Error("Error reading config file: %v", options.AgentConf, err)
			os.Exit(1)
		}

		apiPassword, err := ioutil.ReadFile(options.ApiPasswordPath)
		if err != nil {
			log.Error("Error reading api password file: %v", options.ApiPasswordPath, err)
			os.Exit(1)
		}

		agentConf := map[string]interface{}{}
		err = goyaml.Unmarshal([]byte(yaml), &agentConf)
		if err != nil {
			log.Error("Error reading config file: %v", options.AgentConf, err)
			os.Exit(1)
		}

		clientFactory = func() (*juju.Client, error) {
			//			password := agentConf["apipassword"].(string)
			//			tag := agentConf["tag"].(string)
			//			nonce := agentConf["nonce"].(string)

			password := string(apiPassword)
			tag := "user-admin"
			nonce := ""

			servers := []string{}
			for _, apiaddress := range agentConf["apiaddresses"].([]interface{}) {
				servers = append(servers, apiaddress.(string))
			}

			ca := agentConf["cacert"].(string)
			info := api.Info{
				Addrs:    servers,
				Password: password,
				CACert:   []byte(ca),
				Tag:      tag,
				Nonce:    nonce,
			}

			log.Info("%v", log.AsJson(info))

			return juju.SimpleClientFactory(&info)
		}
	}

	binder.AddProvider(clientFactory)

	bundleStore := bundle.NewBundleStore("templates")
	binder.AddSingleton(bundleStore)

	authenticator := options.Authenticator
	binder.AddSingleton(authenticator)

	apiclient, err := clientFactory()

	// TODO: How would we get the full config "from afar"?
	//confParams := map[string]interface{}{}
	////	confParams["name"] = "jxaas"
	////	confParams["firewall-mode"] = "instance"
	////	confParams["development"] = false
	////
	////	confParams["type"] = "ec2"
	////
	////	confParams["ssl-hostname-verification"] = true
	////	confParams["authorized-keys"] = ""
	////
	//	//		"state-port":                DefaultStatePort,
	//	//		"api-port":                  DefaultAPIPort,
	//	//		"syslog-port":               DefaultSyslogPort,
	//	//		"bootstrap-timeout":         DefaultBootstrapSSHTimeout,
	//	//		"bootstrap-retry-delay":     DefaultBootstrapSSHRetryDelay,
	//	//		"bootstrap-addresses-delay": DefaultBootstrapSSHAddressesDelay,
	//	conf, err := config.New(config.NoDefaults, confParams)
	//	if err != nil {
	//		log.Fatal("Error building Juju config", err)
	//		os.Exit(1)
	//	}
	//	apiclient, err := juju.DirectClientFactory(conf)
	if err != nil {
		log.Fatal("Error building Juju client", err)
		os.Exit(1)
	}

	system := core.NewSystem()

	system.AddBundleType(bundletype.NewMysqlBundleType(bundleStore))
	system.AddBundleType(bundletype.NewElasticsearchBundleType(bundleStore))
	system.AddBundleType(bundletype.NewMongodbBundleType(bundleStore))
	system.AddBundleType(bundletype.NewPgBundleType(bundleStore))
	system.AddBundleType(bundletype.NewMultitenantMysqlBundleType(bundleStore))
	system.AddBundleType(bundletype.NewCassandraBundleType(bundleStore))

	privateUrl := options.PrivateUrl

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
