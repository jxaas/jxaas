package router

import (
	"net/http"
	"github.com/justinsb/gova/log"
)

type Router struct {
	registry RouterRegistry
	listen string
}

func NewRouter(registry RouterRegistry, listen string) *Router {
	self := &Router{}
	self.registry = registry
	self.listen = listen
	return self
}

func (self*Router) httpHandlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Debug("Got http request %v %v", r.Method, r.RequestURI)
	// We'll want to use a new client for every request.
	client := &http.Client{}

	// Tweak the request as appropriate:
	// RequestURI cannot be sent to client
	r.RequestURI = ""

	proxyHost := self.registry.GetBackendForTenant(r.URL.Path)
	if proxyHost == "" {
		http.NotFound(w, r)
		return
	}

	//	URL.Scheme must be lower-case
	log.Debug("scheme %v", r.URL.Scheme)
	log.Debug("path %v", r.URL.Path)
	r.URL.Scheme = "http"
	r.URL.Host = proxyHost

	// And proxy
	resp, err := client.Do(r)
	if err != nil {
		log.Warn("Error serving request", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	statusCode := resp.StatusCode
	w.WriteHeader(statusCode)
	resp.Write(w)
}

func (self*Router) Run() error {
	proxyHandler := http.HandlerFunc(self.httpHandlerFunc)
	err := http.ListenAndServe(self.listen, proxyHandler)
	return err
}
