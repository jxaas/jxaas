package rs

import (
	"net/http"
	"time"

	"bitbucket.org/jsantabarbara/jxaas/inject"
)

type RestServer struct {
	httpServer *http.Server
	injectors  []inject.Injector
}

func NewRestServer() *RestServer {
	self := &RestServer{}
	self.httpServer = &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	self.injectors = []inject.Injector{}
	return self
}

func (self *RestServer) AddInjector(injector inject.Injector) {
	self.injectors = append(self.injectors, injector)
}

func (self *RestServer) AddEndpoint(path string, object interface{}) *RestEndpointHandler {
	endpoint := newRestEndpoint(self, path, object)
	return endpoint
}

func (self *RestServer) ListenAndServe() error {
	return self.httpServer.ListenAndServe()
}
