package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	//	"launchpad.net/gnuflag"
	//
	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/constraints"
	//	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/state/api"
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

var connectionError = `Unable to connect to environment "%s".
Please check your credentials or use 'juju bootstrap' to create a new environment.

Error details:
%v
`

type CharmsEndpoint struct {
}

type Instance struct {
	Id string

	Units map[string]*Unit
}

type Unit struct {
	Id string

	PublicAddress string

	Status string
}

func MapToUnit(id string, api *api.UnitStatus) *Unit {
	unit := &Unit{}
	unit.Id = id
	unit.PublicAddress = api.PublicAddress
	unit.Status = string(api.AgentState)
	return unit
}

func MapToInstance(id string, api *api.ServiceStatus) *Instance {
	instance := &Instance{}
	instance.Id = id
	instance.Units = make(map[string]*Unit)
	for key, unit := range api.Units {
		instance.Units[key] = MapToUnit(key, &unit)
	}
	return instance
}

func (self *CharmsEndpoint) List() ([]*Instance, error) {
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

	instances := make([]*Instance, 0)
	for key, state := range status.Services {
		fmt.Printf("%v => %v\n\n", key, state)
		instance := MapToInstance(key, &state)

		instances = append(instances, instance)
	}

	fmt.Printf("%v", status)

	return instances, nil
	//
	//	result := formatStatus(status)
	//
	//	return c.out.Write(ctx, result), nil
}

func (self *CharmsEndpoint) Find(id string) (*Instance, error) {
	//	return "Hello world"
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	patterns := make([]string, 1)
	patterns[0] = id
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

	state, found := status.Services[id]
	if !found {
		return nil, HttpError(http.StatusNotFound)
	}

	return MapToInstance(id, &state), nil
}

func (self *CharmsEndpoint) Create() (*Instance, error) {
	//	return "Hello world"
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	//	curl, err := charm.InferURL(c.CharmName, conf.DefaultSeries())
	//	if err != nil {
	//		return err
	//	}
	//	repo, err := charm.InferRepository(curl, ctx.AbsPath(c.RepoPath))
	//	if err != nil {
	//		return err
	//	}
	//
	//	repo = config.SpecializeCharmRepo(repo, conf)
	//
	//	curl, err = addCharmViaAPI(client, ctx, curl, repo)
	//	if err != nil {
	//		return err
	//	}

	//
	//	charmInfo, err := client.CharmInfo(curl.String())
	//	if err != nil {
	//		return err
	//	}

	numUnits := 1

	serviceName := "service" + strconv.Itoa(rand.Int())

	//	if serviceName == "" {
	//		serviceName = charmInfo.Meta.Name
	//	}

	var configYAML []byte
	//	if c.Config.Path != "" {
	//		configYAML, err = c.Config.Read(ctx)
	//		if err != nil {
	//			return err
	//		}
	//	}
	var constraints constraints.Value
	var toMachineSpec string

	charmUrl := "cs:precise/mysql-38"

	err = apiclient.ServiceDeploy(
		charmUrl,
		serviceName,
		numUnits,
		string(configYAML),
		constraints,
		toMachineSpec,
	)

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

	return self.Find(serviceName)
}

type HttpResponse struct {
	Status  int
	Content []byte
	Headers map[string]string
}

func (self *CharmsEndpoint) Delete(serviceName string) (*HttpResponse, error) {
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	err = apiclient.ServiceDestroy(serviceName)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{Status: http.StatusAccepted}, nil
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

	requestUri := req.RequestURI
	if len(requestUri) > 0 && requestUri[0] == '/' {
		requestUri = requestUri[1:]
	}

	if len(requestUri) > 0 && requestUri[len(requestUri)-1] == '/' {
		requestUri = requestUri[:len(requestUri)-1]
	}

	pathComponents := strings.Split(requestUri, "/")

	log.Info("Path components: %v", pathComponents)

	var err error

	methodName := ""

	args := make([]reflect.Value, 1)
	args[0] = o

	if req.Method == "GET" {
		prefixComponents := 1

		if len(pathComponents) == prefixComponents {
			methodName = "List"
		} else if len(pathComponents) == (prefixComponents + 1) {
			methodName = "Find"
			args = append(args, reflect.ValueOf(pathComponents[prefixComponents]))
		} else {
			log.Debug("Could not match path: %v (len=%v)", pathComponents, len(pathComponents))

			err = HttpError(http.StatusNotFound)
		}
	}

	if req.Method == "DELETE" {
		prefixComponents := 1

		if len(pathComponents) == (prefixComponents + 1) {
			methodName = "Delete"
			args = append(args, reflect.ValueOf(pathComponents[prefixComponents]))
		} else {
			log.Debug("Could not match path: %v (len=%v)", pathComponents, len(pathComponents))

			err = HttpError(http.StatusNotFound)
		}
	}
	if req.Method == "POST" {
		prefixComponents := 1

		if len(pathComponents) == prefixComponents {
			methodName = "Create"
			//args = append(args, reflect.ValueOf(pathComponents[prefixComponents]))
		} else {
			log.Debug("Could not match path: %v (len=%v)", pathComponents, len(pathComponents))

			err = HttpError(http.StatusNotFound)
		}
	}

	var method reflect.Method
	if err == nil {
		var found bool
		method, found = self.ptrT.MethodByName(methodName)
		if !found {
			log.Debug("Method not found: %v", methodName)

			err = HttpError(http.StatusNotFound)
		}
	}

	var val reflect.Value

	if err == nil {
		var out []reflect.Value
		out = method.Func.Call(args)
		//		fmt.Fprintf(w, "Returned %v", out)

		if len(out) >= 2 {
			errValue := out[1]
			if !errValue.IsNil() {
				var ok bool
				log.Debug("Got error value: %v", errValue)
				err, ok = errValue.Interface().(error)
				if !ok {
					log.Warn("Unable to cast value to error: %v", errValue)
					err = HttpError(http.StatusInternalServerError)
				}
			}
		}

		if err == nil && len(out) > 0 {
			// TODO: Don't assume position 0
			val = out[0]
			if val.IsNil() {
				err = HttpError(http.StatusNotFound)
			}
		}
	}

	if err == nil {
		if !val.IsValid() || val.IsNil() {
			err = HttpError(http.StatusNotFound)
		}
	}

	var response *HttpResponse
	if err == nil {
		var ok bool
		response, ok = val.Interface().(*HttpResponse)
		if !ok {
			data, _ := json.Marshal(val.Interface())
			response = &HttpResponse{Status: http.StatusOK}
			response.Content = data
			response.Headers = make(map[string]string)
			response.Headers["content-type"] = "application/json; charset=utf-8"
		}

		if response == nil {
			log.Warn("Unable to build response for %v", val)
			err = HttpError(http.StatusInternalServerError)
		}
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

func main() {
	juju.InitJujuHome()

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	NewRestEndpoint("/charm/", (*CharmsEndpoint)(nil))

	log.Fatal("Error serving HTTP", s.ListenAndServe())
}
