package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"reflect"
	"time"

	//	"launchpad.net/gnuflag"
	//
	"launchpad.net/juju-core/cmd"
	//	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/juju"
	//	"launchpad.net/juju-core/state/api"
	//	"launchpad.net/juju-core/state/api/params"
	//	"launchpad.net/juju-core/state/statecmd"
)

var connectionError = `Unable to connect to environment "%s".
Please check your credentials or use 'juju bootstrap' to create a new environment.

Error details:
%v
`

type CharmsEndpoint struct {
}

type Instance struct {
	Id string
}

func (self *CharmsEndpoint) Get() ([]Instance, error) {
	//	return "Hello world"
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	patterns := make([]string, 0)

	status, err := apiclient.Status(patterns)

	//	if params.IsCodeNotImplemented(err) {
	//		logger.Infof("Status not supported by the API server, " +
	//			"falling back to 1.16 compatibility mode " +
	//			"(direct DB access)")
	//		status, err = c.getStatus1dot16()
	//	}
	// Display any error, but continue to print status if some was returned
	if err != nil {
		return nil, err
	}

	instances := make([]Instance, 0)
	for key, state := range status.Services {
		fmt.Printf("%v => %v\n\n", key, state)
		instance := Instance{}
		instance.Id = key

		instances = append(instances, instance)
	}

	fmt.Printf("%v", status)

	return instances, nil
	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil
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

func (self *RestEndpointHandler) httpHandler(res http.ResponseWriter, req *http.Request) {
	o := reflect.New(self.structT)

	if req.Method == "GET" {
		method, found := self.ptrT.MethodByName("Get")
		if !found {
			// TODO: 404
			panic("Method not found")
		}
		args := make([]reflect.Value, 1)
		args[0] = o
		out := method.Func.Call(args)
		//		fmt.Fprintf(w, "Returned %v", out)

		// TODO: Handle error
		// TODO: Don't assume position 0
		data, _ := json.Marshal(out[0].Interface())
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Write(data)
	}

	//	fmt.Fprintf(w, "Hello, %v", html.EscapeString(req.URL.Path))
	//	fmt.Fprintf(w, "Hello, %v", self.ptrT)
}

func main() {
	juju.InitJujuHome()

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	NewRestEndpoint("/charm/", (*CharmsEndpoint)(nil))

	log.Fatal(s.ListenAndServe())
}
