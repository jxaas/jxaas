package router

import (
	"io"
	"net/http"
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"

	"reflect"
)

type Router struct {
	registry RouterRegistry
	listen   string
}

func NewRouter(registry RouterRegistry, listen string) *Router {
	self := &Router{}
	self.registry = registry
	self.listen = listen
	return self
}

// http.Header is a map, which is a ref-type
func copyHeaders(src http.Header, dest http.Header) {
	for key, values := range src {
		for _, value := range values {
			dest.Add(key, value)
		}
	}
}

func (self*Router) httpHandlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Debug("Got http request %v %v", r.Method, r.RequestURI)

	// Tweak the request as appropriate:
	// RequestURI cannot be sent to client
	r.RequestURI = ""

	proxyHost := ""

	path := r.URL.Path
	if path[0] == '/' {
		path = path[1:]
	}

	tokens := strings.Split(path, "/")
	log.Debug("URL tokens: %v", tokens)
	if len(tokens) >= 3 {
		app := tokens[0]
		if app == "xaas" {
			tenant := tokens[1]
			if len(tokens) == 3 && tokens[2] == "services" {
				self.listServices(tenant, r, w)
				return
			} else if len(tokens) >= 4 {
				service := tokens[3]
				proxyHost = self.registry.GetBackendForTenant(service, tenant)
			}
		}
	}

	if proxyHost == "" {
		http.NotFound(w, r)
		return
	}

	//	URL.Scheme must be lower-case
	//	log.Debug("scheme %v", r.URL.Scheme)
	//	log.Debug("path %v", r.URL.Path)
	r.URL.Scheme = "http"
	r.URL.Host = proxyHost

	log.Debug("Proxying to %v", r.URL)

	// Proxy the request
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Warn("Error serving request", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	copyHeaders(resp.Header, w.Header())

	statusCode := resp.StatusCode
	w.WriteHeader(statusCode)

	io.Copy(w, resp.Body)
}

func (self*Router) Run() error {
	proxyHandler := http.HandlerFunc(self.httpHandlerFunc)
	err := http.ListenAndServe(self.listen, proxyHandler)
	return err
}

func (self * Router) listServices(tenant string, req *http.Request, res http.ResponseWriter) {
	bundles, err := self.registry.ListServicesForTenant(tenant)
	if err != nil {
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}
	sendJsonResponse(bundles, req, res)
}

func sendJsonResponse(data interface{}, req *http.Request, res http.ResponseWriter) {
	writer := rs.NewJsonMessageBodyWriter()
	t := reflect.TypeOf(data)
	writer.Write(data, t, req, res)
}


