package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	//	"launchpad.net/gnuflag"
	//
	"github.com/justinsb/gova/log"

	//	"launchpad.net/juju-core/instance"

	//	"launchpad.net/juju-core/state/api"
	//	"launchpad.net/juju-core/state/api/params"
	//	"launchpad.net/juju-core/state/statecmd"
)

type HttpErrorObject struct {
	Status  int
	Message string
}

func HttpError(status int) *HttpErrorObject {
	e := &HttpErrorObject{}
	e.Status = status
	return e
}

func (self *HttpErrorObject) Error() string {
	return ""
}

type HttpResponse struct {
	Status  int
	Content []byte
	Headers map[string]string
}

type RestEndpointHandler struct {
	path string

	ptrT    reflect.Type
	structT reflect.Type
}

func NewRestEndpoint(path string, object interface{}) *RestEndpointHandler {
	self := &RestEndpointHandler{}

	self.ptrT = reflect.TypeOf(object)
	self.structT = self.ptrT.Elem()
	self.path = path

	http.HandleFunc(path, self.httpHandler)

	return self
}

func parseReturn(out []reflect.Value) (reflect.Value, error) {
	var value reflect.Value
	var err error

	if len(out) >= 2 {
		// TODO: Don't assume position 1?
		errValue := out[1]
		if !errValue.IsNil() {
			var ok bool
			log.Debug("Got error value: %v", errValue)
			err, ok = errValue.Interface().(error)
			if !ok {
				err = fmt.Errorf("Unable to cast value to error")
			}
		}
	}

	if err == nil && len(out) > 0 {
		// TODO: Don't assume position 0
		value = out[0]

		if !value.IsValid() {
			value = reflect.ValueOf(nil)
		}
	}

	return value, err
}

func (self *RestEndpointHandler) resolveEndpoint(res http.ResponseWriter, req *http.Request) (*reflect.Value, error) {
	requestUri := req.RequestURI
	suffix := requestUri[len(self.path):]

	if len(suffix) > 0 && suffix[0] == '/' {
		suffix = suffix[1:]
	}

	if len(suffix) > 0 && suffix[len(suffix)-1] == '/' {
		suffix = suffix[:len(suffix)-1]
	}

	var err error

	endpoint := reflect.New(self.structT)

	if suffix != "" {
		pathComponents := strings.Split(suffix, "/")

		log.Info("Path components:  %v", pathComponents)

		for _, pathComponent := range pathComponents {
			itemMethod := endpoint.MethodByName("Item")
			if !itemMethod.IsValid() {
				log.Debug("Items method not found")

				return nil, nil
			}

			in := []reflect.Value{reflect.ValueOf(pathComponent)}

			out := itemMethod.Call(in)
			endpoint, err = parseReturn(out)
			if err != nil {
				return nil, err
			}
			if endpoint.IsNil() {
				return nil, nil
			}
		}
	}

	return &endpoint, nil
}

func (self *RestEndpointHandler) makeResponse(val reflect.Value) (*HttpResponse, error) {
	var ok bool
	response, ok := val.Interface().(*HttpResponse)
	if !ok {
		data, _ := json.Marshal(val.Interface())
		response = &HttpResponse{Status: http.StatusOK}
		response.Content = data
		response.Headers = make(map[string]string)
		response.Headers["content-type"] = "application/json; charset=utf-8"
	}

	if response == nil {
		log.Warn("Unable to build response for %v", val)
		return nil, fmt.Errorf("Unable to build response")
	}

	return response, nil
}

func (self *RestEndpointHandler) httpHandler(res http.ResponseWriter, req *http.Request) {
	endpoint, err := self.resolveEndpoint(res, req)

	args := make([]reflect.Value, 0)

	if endpoint == nil {
		err = HttpError(http.StatusNotFound)
	}

	var method reflect.Value

	if err == nil {
		httpMethod := req.Method
		methodName := "Http" + httpMethod[0:1] + strings.ToLower(httpMethod[1:])

		method = endpoint.MethodByName(methodName)
		if !method.IsValid() {
			log.Debug("Method not found: %v", methodName)

			err = HttpError(http.StatusNotFound)
		}
	}

	var val reflect.Value

	if err == nil {
		var out []reflect.Value
		out = method.Call(args)
		//		fmt.Fprintf(w, "Returned %v", out)

		val, err = parseReturn(out)
	}

	if err == nil {
		if val.IsNil() {
			err = HttpError(http.StatusNotFound)
		}
	}

	var response *HttpResponse

	if err == nil {
		response, err = self.makeResponse(val)
	}

	if err == nil && response != nil {
		if response.Headers != nil {
			for name, value := range response.Headers {
				res.Header().Set(name, value)
			}
		}

		res.WriteHeader(response.Status)

		res.Write(response.Content)
	} else if err == nil && response == nil {
		res.WriteHeader(http.StatusNoContent)
	} else {
		httpError, ok := err.(*HttpErrorObject)
		if !ok {
			log.Warn("Internal error serving request", err)
			httpError = HttpError(http.StatusInternalServerError)
		}

		status := httpError.Status
		message := httpError.Message
		if message == "" {
			message = http.StatusText(status)
			if message == "" {
				message = "Error"
			}
		}

		http.Error(res, message, status)
	}

	//	fmt.Fprintf(w, "Hello, %v", html.EscapeString(req.URL.Path))
	//	fmt.Fprintf(w, "Hello, %v", self.ptrT)
}
