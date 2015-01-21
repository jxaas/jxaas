package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"fmt"
	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/router"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	options := router.GetOptions()
	if options == nil {
		log.Fatal("Error reading options")
		os.Exit(1)
	}

	args := flag.Args()
	command := ""
	if len(args) != 0 {
		command = args[0]
	}

	if command == "set-service-backend" {
		if len(args) != 3 {
			fmt.Println("Syntax: set-service-backend <service> <backend>")
			os.Exit(1)
		}
		options.Registry.SetBackendForService(args[1], args[2])
		return
	}

	if command == "list-service-backends" {
		if len(args) != 1 {
			fmt.Println("Syntax: list-service-backends")
			os.Exit(1)
		}
		services, err := options.Registry.ListServices()
		if err != nil {
			log.Fatal("Error reading services", err)
			os.Exit(2)
		}
		for _, service := range services {
			backend := options.Registry.GetBackendForTenant(service, "")
			fmt.Println(service, "\t", backend)
		}
		return
	}

	if command != "" {
		fmt.Println("Unknown command:", command)
	}
	fmt.Println("Valid commands:")
	fmt.Println("  set-service-backend")
	os.Exit(1)
}
