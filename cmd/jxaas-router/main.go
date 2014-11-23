package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/justinsb/gova/log"
	"github.com/jxaas/jxaas/router"
	"fmt"
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

	if command == "set" {
		if len(args) != 3 {
			fmt.Println("Syntax: add <tenant> <backend>");
			os.Exit(1)
		}
		options.Registry.SetBackendForTenant(args[1], args[2])
		return
	}

	if command != "" {
		fmt.Println("Unknown command:", command)
	}
	fmt.Println("Valid commands:");
	fmt.Println("  set");
	os.Exit(1)
}
