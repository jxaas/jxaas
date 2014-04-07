package rs

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type MessageBodyReader interface {
	IsReadable(t reflect.Type, req *http.Request, mediaType *MediaType) bool
	Read(t reflect.Type, req *http.Request, mediaType *MediaType) (interface{}, error)
}

type JsonMessageBodyReader struct {
}

func NewJsonMessageBodyReader() *JsonMessageBodyReader {
	self := &JsonMessageBodyReader{}
	return self
}

func (self *JsonMessageBodyReader) IsReadable(t reflect.Type, req *http.Request, mediaType *MediaType) bool {
	if mediaType.Is("application/json") {
		return true
	}

	return false
}

func (self *JsonMessageBodyReader) Read(t reflect.Type, req *http.Request, mediaType *MediaType) (interface{}, error) {
	body := req.Body
	defer body.Close()

	decoder := json.NewDecoder(req.Body)

	m := reflect.New(t).Interface()

	err := decoder.Decode(&m)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return m, nil
}
