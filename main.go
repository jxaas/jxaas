package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"reflect"
	"time"
)

type CharmsEndpoint struct {
}

func (self *CharmsEndpoint) Get() string {
	return "Hello world"
}


type RestEndpointHandler struct {
	ptrT    reflect.Type
	structT reflect.Type
}

func NewRestEndpoint(path string, object interface{}) *RestEndpointHandler {
	self := &RestEndpointHandler{}

	self.ptrT = reflect.TypeOf(object)
	self.structT = self.ptrT.Elem()

	http.HandleFunc(path, self.httpHandler)

	return self
}

func (self *RestEndpointHandler) httpHandler(w http.ResponseWriter, req *http.Request) {
	o := reflect.New(self.structT)

	if req.Method == "GET" {
		method, found := self.ptrT.MethodByName("Get")
		if !found {
			// TOOD: 404
			panic("Method not found")
		}
		args := make([]reflect.Value, 1)
		args[0] = o
		out := method.Func.Call(args)
		fmt.Fprintf(w, "Returned %v", out)
	}

	//	fmt.Fprintf(w, "Hello, %v", html.EscapeString(req.URL.Path))
	//	fmt.Fprintf(w, "Hello, %v", self.ptrT)
}

func main() {
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.HandleFunc("/bar/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	NewRestEndpoint("/charm/", (*CharmsEndpoint)(nil))

	log.Fatal(s.ListenAndServe())
}
