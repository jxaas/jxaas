package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/auth"
)

var (
	flagAgentConf       = flag.String("c", "", "Agent conf file")
	flagApiPasswordPath = flag.String("p", "", "File containing API password")
	flagPrivateUrl      = flag.String("private", "", "Private URL")
	flagKeystoneUrl     = flag.String("openstack", "http://127.0.0.1:5000/v2.0", "URL for OpenStack Identity service")
	flagAuth            = flag.String("auth", "development", "Authentication plugin to use")
	flagCfTenantId      = flag.String("cf-tenant", "cf", "TenantId to use for cloudfoundry services")
	flagListenAddress   = flag.String("listen", ":8080", "Address on which to listen")
)

type Options struct {
	AgentConf       string
	ApiPasswordPath string
	PrivateUrl      string
	Authenticator   auth.Authenticator
	CfTenantId      string
	ListenAddress   string
}

func localIP() (net.IP, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, netInterface := range netInterfaces {
		addresses, err := netInterface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, address := range addresses {
			ipnet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address
				continue
			}
			log.Info("Chose local IP: %v", v4)
			return v4, nil
		}
	}
	return nil, errors.New("Cannot find local IP address")
}

func GetOptions() *Options {
	flag.Parse()

	self := &Options{}

	self.AgentConf = *flagAgentConf
	self.ApiPasswordPath = *flagApiPasswordPath

	self.CfTenantId = *flagCfTenantId
	self.ListenAddress = *flagListenAddress

	host, port, err := net.SplitHostPort(self.ListenAddress)
	if err != nil {
		log.Warn("Cannot parse listen address: %v", self.ListenAddress)
		return nil
	}
	var portNum int
	if port == "" {
		portNum = 8080
	} else {
		portNum, err = net.LookupPort("tcp", port)
		if err != nil {
			log.Warn("Cannot resolve port: %v", port)
			return nil
		}
	}

	privateUrl := *flagPrivateUrl
	if privateUrl == "" {
		privateHost := host
		if privateHost == "" {
			ip, err := localIP()
			if err != nil {
				log.Warn("Error finding local IP", err)
				return nil
			}
			privateHost = ip.String()
		}

		privateUrl = fmt.Sprintf("http://%v:%v/xaasprivate", privateHost, portNum)
		log.Info("Chose private url: %v", privateUrl)
	}
	self.PrivateUrl = privateUrl

	authMode := *flagAuth
	authMode = strings.TrimSpace(authMode)
	authMode = strings.ToLower(authMode)
	if authMode == "openstack" {
		keystoneUrl := *flagKeystoneUrl
		self.Authenticator = auth.NewOpenstackMultiAuthenticator(keystoneUrl)
	} else if authMode == "development" {
		self.Authenticator = auth.NewDevelopmentAuthenticator()
	} else {
		log.Warn("Unknown authentication mode: %v", authMode)
		return nil
	}

	return self
}
