package main

import "flag"

var (
	flagAgentConf = flag.String("c", "", "Agent conf file")
)

type Options struct {
	AgentConf string
}

func GetOptions() *Options {
	flag.Parse()

	self := &Options{}

	self.AgentConf = *flagAgentConf

	return self
}
