package rs

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
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

	valueT := t
	pointerDepth := 0
	for valueT.Kind() == reflect.Ptr {
		valueT = valueT.Elem()
		pointerDepth++
	}

	msg := reflect.New(valueT)

	err := decoder.Decode(msg.Interface())
	if err != nil { // && err != io.EOF {
		return nil, err
	}

	log.Warn("PointerDepth: %v, type %v", pointerDepth, t)

	// reflect.New returns a pointer, so we start at 1
	for i := 1; i < pointerDepth; i++ {
		assert.That(msg.CanAddr())
		msg = msg.Addr()
	}

	assert.Equal(msg.Type(), t)

	return msg.Interface(), nil
}
