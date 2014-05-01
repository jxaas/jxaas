package main

import "flag"

var (
	flagAgentConf       = flag.String("c", "", "Agent conf file")
	flagApiPasswordPath = flag.String("p", "", "File containing API password")
)

type Options struct {
	AgentConf       string
	ApiPasswordPath string
}

func GetOptions() *Options {
	flag.Parse()

	self := &Options{}

	self.AgentConf = *flagAgentConf
	self.ApiPasswordPath = *flagApiPasswordPath
	return self
}
