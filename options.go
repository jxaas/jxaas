package main

import "flag"

var (
	flagAgentConf   = flag.String("c", "", "Agent conf file")
	flagApiPassword = flag.String("p", "", "API password")
)

type Options struct {
	AgentConf   string
	ApiPassword string
}

func GetOptions() *Options {
	flag.Parse()

	self := &Options{}

	self.AgentConf = *flagAgentConf
	self.ApiPassword = *flagApiPassword
	return self
}
