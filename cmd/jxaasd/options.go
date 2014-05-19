package main

import (
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/justinsb/gova/log"
)

var (
	flagAgentConf       = flag.String("c", "", "Agent conf file")
	flagApiPasswordPath = flag.String("p", "", "File containing API password")
	flagPrivateUrl      = flag.String("private", "", "Private URL")
)

type Options struct {
	AgentConf       string
	ApiPasswordPath string
	PrivateUrl      string
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

	privateUrl := *flagPrivateUrl
	if privateUrl == "" {
		ip, err := localIP()
		if err != nil {
			log.Warn("Error finding local IP", err)
			return nil
		}

		privateUrl = fmt.Sprintf("http://%v:8080/xaasprivate", ip)
		log.Info("Chose private url: %v", privateUrl)
	}

	self.PrivateUrl = privateUrl

	return self
}
