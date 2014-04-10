package rs

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type MessageBodyWriter interface {
	IsWritable(t reflect.Type, req *http.Request, mediaType *MediaType) bool
	Write(obj interface{}, t reflect.Type, req *http.Request, res http.ResponseWriter) error
}

type NoResponseMessageBodyWriter struct {
}

func (self *NoResponseMessageBodyWriter) IsWritable(t reflect.Type, req *http.Request, mediaType *MediaType) bool {
	return false
}

func (self *NoResponseMessageBodyWriter) Write(o interface{}, t reflect.Type, req *http.Request, res http.ResponseWriter) error {
	return nil
}

type JsonMessageBodyWriter struct {
}

func NewJsonMessageBodyWriter() *JsonMessageBodyWriter {
	self := &JsonMessageBodyWriter{}
	return self
}

func (self *JsonMessageBodyWriter) IsWritable(t reflect.Type, req *http.Request, mediaType *MediaType) bool {
	if mediaType.Is("application/json") {
		return true
	}

	return false
}

func (self *JsonMessageBodyWriter) Write(o interface{}, t reflect.Type, req *http.Request, res http.ResponseWriter) error {
	encoder := json.NewEncoder(res)

	err := encoder.Encode(o)
	if err != nil {
		return err
	}

	return nil
}
