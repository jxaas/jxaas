package rs

import (
	"net/http"
	"reflect"
	"time"

	"github.com/justinsb/gova/assert"

	"github.com/jxaas/jxaas/inject"
)

type RestServer struct {
	httpServer *http.Server
	injector   inject.Injector

	readers []MessageBodyReader
	writers []MessageBodyWriter

	defaultMediaType *MediaType
}

func NewRestServer() *RestServer {
	self := &RestServer{}
	self.httpServer = &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	self.readers = []MessageBodyReader{}
	self.writers = []MessageBodyWriter{}

	var err error
	self.defaultMediaType, err = ParseMediaType("application/json")
	assert.That(err == nil)

	return self
}

func (self *RestServer) WithInjector(injector inject.Injector) {
	self.injector = injector
}

func (self *RestServer) AddEndpoint(path string, object interface{}) *RestEndpointHandler {
	endpoint := newRestEndpoint(self, path, object)
	return endpoint
}

func (self *RestServer) readMessageBody(t reflect.Type, req *http.Request, mediaType *MediaType) (interface{}, error) {
	for _, mbr := range self.readers {
		if !mbr.IsReadable(t, req, mediaType) {
			continue
		}

		return mbr.Read(t, req, mediaType)
	}

	return nil, nil
}

func (self *RestServer) findMessageBodyWriter(object interface{}, req *http.Request, mediaType *MediaType) MessageBodyWriter {
	t := reflect.TypeOf(object)

	for _, mbw := range self.writers {
		if !mbw.IsWritable(t, req, mediaType) {
			continue
		}

		return mbw
	}

	return nil
}

func (self *RestServer) AddReader(mbr MessageBodyReader) {
	self.readers = append(self.readers, mbr)
}

func (self *RestServer) AddWriter(mbw MessageBodyWriter) {
	self.writers = append(self.writers, mbw)
}

func (self *RestServer) ListenAndServe() error {
	return self.httpServer.ListenAndServe()
}
