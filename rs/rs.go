package rs

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/log"
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

func ErrNotFound() *HttpErrorObject {
	err := HttpError(http.StatusNotFound)
	return err
}

type HttpResponse struct {
	Status  int
	Content interface{}
	Headers map[string]string
}

type RestEndpointHandler struct {
	server *RestServer
	path   string

	ptrT    reflect.Type
	structT reflect.Type
}

func newRestEndpoint(server *RestServer, path string, object interface{}) *RestEndpointHandler {
	self := &RestEndpointHandler{}

	self.server = server
	self.path = path

	self.ptrT = reflect.TypeOf(object)
	self.structT = self.ptrT.Elem()

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

			methodName := "Item" + strings.ToUpper(pathComponent[:1]) + strings.ToLower(pathComponent)[1:]
			method := endpoint.MethodByName(methodName)
			if !method.IsValid() {
				method = endpoint.MethodByName("Item")
				if !method.IsValid() {
					log.Debug("Items method not found (nor %v)", methodName)
					return nil, nil
				}
			}

			injector := self.server.injector
			methodType := method.Type()
			numIn := method.Type().NumIn()
			args := make([]reflect.Value, numIn, numIn)
			for i := 0; i < numIn; i++ {
				argType := methodType.In(i)
				if argType.Kind() == reflect.String {
					args[i] = reflect.ValueOf(pathComponent)
				} else {
					v, err := injector.Get(argType)
					if err != nil {
						log.Warn("Error injecting argument of type: %v", argType, err)
						return nil, err
					}
					args[i] = reflect.ValueOf(v)
				}
			}

			out := method.Call(args)
			endpoint, err = parseReturn(out)
			if err != nil {
				return nil, err
			}
			if endpoint.IsNil() {
				return nil, nil
			}
			if endpoint.Kind() == reflect.Interface {
				endpoint = endpoint.Elem()
			}
		}
	}

	return &endpoint, nil
}

func (self *RestEndpointHandler) makeResponse(val reflect.Value) (*HttpResponse, error) {
	var ok bool
	var response *HttpResponse

	if val.IsNil() {
		response = &HttpResponse{Status: http.StatusNoContent}
		response.Content = nil
	} else {
		response, ok = val.Interface().(*HttpResponse)
		if !ok {
			response = &HttpResponse{Status: http.StatusOK}
			response.Content = val.Interface()
		}
	}

	if response == nil {
		log.Warn("Unable to build response for %v", val)
		return nil, fmt.Errorf("Unable to build response")
	}

	return response, nil
}

func getMediaType(req *http.Request) (*MediaType, error) {
	v := req.Header.Get("Content-Type")
	if v == "" {
		return nil, nil
	}

	mediaType, err := ParseMediaType(v)
	if err != nil {
		log.Warn("Error parsing mime type: %v", v, err)
		return nil, err
	}

	return mediaType, nil
}

func (self *RestEndpointHandler) buildArg(res http.ResponseWriter, req *http.Request, t reflect.Type) (interface{}, error) {
	v, err := self.server.injector.Get(t)
	if err == nil && v != nil {
		return v, nil
	}

	// TODO: Fail if two args...

	// TODO: Only if has content?
	mediaType, err := getMediaType(req)
	if err != nil {
		return nil, err
	}

	if mediaType == nil {
		// Go does have a function to guess the media type, but that seems risky
		// Instead, use a fixed default
		mediaType = self.server.defaultMediaType
	}

	v, err = self.server.readMessageBody(t, req, mediaType)
	if err != nil {
		if err == io.EOF {
			log.Debug("Error reading message body (EOF)")
		} else {
			log.Debug("Error reading message body", err)
		}
		err = HttpError(http.StatusBadRequest)
		return nil, err
	}

	if v == nil && err == nil {
		err = HttpError(http.StatusUnsupportedMediaType)
		return nil, err
	}

	if v != nil {
		assert.Equal(reflect.TypeOf(v), t)
		return v, nil
	}

	log.Warn("Unable to bind parameter: %v", t)
	return nil, fmt.Errorf("Unable to bind parameter: %v", t)
}

func (self *RestEndpointHandler) buildArgs(res http.ResponseWriter, req *http.Request, method *reflect.Value) ([]reflect.Value, error) {
	numIn := method.Type().NumIn()
	args := make([]reflect.Value, numIn)
	methodType := method.Type()
	for i := 0; i < methodType.NumIn(); i++ {
		val, err := self.buildArg(res, req, methodType.In(i))
		if err != nil {
			return nil, err
		}
		if val != nil {
			args[i] = reflect.ValueOf(val)
		}
	}

	return args, nil
}

func (self *RestEndpointHandler) httpHandler(res http.ResponseWriter, req *http.Request) {
	requestUrl := req.URL
	requestMethod := req.Method

	log.Debug("%v %v", requestMethod, requestUrl)

	endpoint, err := self.resolveEndpoint(res, req)

	if endpoint == nil {
		err = HttpError(http.StatusNotFound)
	}

	var method reflect.Value

	if err == nil {
		httpMethod := req.Method
		methodName := "Http" + httpMethod[0:1] + strings.ToLower(httpMethod[1:])

		method = endpoint.MethodByName(methodName)
		if !method.IsValid() {
			log.Debug("Method not found: %v on %v", methodName, endpoint.Type())

			err = HttpError(http.StatusNotFound)
		}
	}

	var args []reflect.Value
	if err == nil {
		args, err = self.buildArgs(res, req, &method)
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
	var mbw MessageBodyWriter

	if err == nil {
		response, err = self.makeResponse(val)
	}
	if err == nil {
		assert.That(response != nil)

		if response.Headers == nil {
			response.Headers = make(map[string]string)
		}

		if response.Content == nil {
			mbw = &NoResponseMessageBodyWriter{}
		} else {
			contentType := response.Headers["content-type"]
			if contentType == "" {
				contentType = "application/json; charset=utf-8"
				response.Headers["content-type"] = contentType
			}

			var mediaType *MediaType
			if contentType != "" {
				mediaType, err = ParseMediaType(contentType)
			}

			if err == nil {
				assert.That(mediaType != nil)

				mbw = self.server.findMessageBodyWriter(response.Content, req, mediaType)
				if mbw == nil {
					log.Warn("Unable to find media type: %v", contentType)
					err = HttpError(http.StatusUnsupportedMediaType)
				}
			}
		}
	}

	if err == nil {
		assert.That(response != nil)
		assert.That(mbw != nil)

		log.Info("%v %v %v", response.Status, requestMethod, requestUrl)

		if response.Headers != nil {
			for name, value := range response.Headers {
				res.Header().Set(name, value)
			}
		}

		res.WriteHeader(response.Status)

		mbw.Write(response.Content, reflect.TypeOf(response.Content), req, res)
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

		log.Info("%v %v %v", status, requestMethod, requestUrl)

		http.Error(res, message, status)
	}
}
