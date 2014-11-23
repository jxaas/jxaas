package main

import (
	"math/rand"
	"os"
	"time"

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

	r := router.NewRouter(options.Registry, options.Listen)
	err := r.Run()
	if err != nil {
		log.Fatal("Error running router", err)
	}
}
