package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
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

var connectionError = `Unable to connect to environment "%s".
Please check your credentials or use 'juju bootstrap' to create a new environment.

Error details:
%v
`

type EndpointCharms struct {
}

func (self *EndpointCharms) Item(key string) *EndpointCharm {
	charm := &EndpointCharm{}
	charm.Key = key
	return charm
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

func (self *EndpointCharms) HttpGet() ([]*Instance, error) {
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

func (self *EndpointCharms) HttpPost() (*Instance, error) {
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

	return self.Item(serviceName).HttpGet()
}

type EndpointCharm struct {
	Key string
}

func (self *EndpointCharm) Item(key string) (interface{}, error) {
	// TODO: Auto-discover using struct fields
	if key == "log" {
		log := &EndpointLog{}
		log.Parent = self
		return log, nil
	} else if key == "metrics" {
		log := &EndpointMetrics{}
		log.Parent = self
		return log, nil
	} else {
		return nil, nil
	}
}

func (self *EndpointCharm) HttpGet() (*Instance, error) {
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	// TODO: Is this efficient?  Any direct just-this-service call?
	patterns := make([]string, 1)
	patterns[0] = self.Key
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

	state, found := status.Services[self.Key]
	if !found {
		return nil, HttpError(http.StatusNotFound)
	}

	log.Debug("Service state: %v", state)

	return MapToInstance(self.Key, &state), nil
}

func (self *EndpointCharm) HttpDelete() (*HttpResponse, error) {
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}
	defer apiclient.Close()

	serviceName := self.Key

	err = apiclient.ServiceDestroy(serviceName)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{Status: http.StatusAccepted}, nil
}

type EndpointLog struct {
	Parent *EndpointCharm
}

type Lines struct {
	Line []string
}

func (self *EndpointLog) HttpGet() (*Lines, error) {
	service := self.Parent.Key

	// TODO: Inject
	logStore := &JujuLogStore{}
	logStore.basedir = "/var/log/juju-justinsb-local/"

	// TODO: Expose units?
	unitId := 0

	logfile, err := logStore.ReadLog(service, unitId)
	if err != nil {
		log.Warn("Error reading log: %v", unitId, err)
		return nil, err
	}
	if logfile == nil {
		log.Warn("Log not found: %v", unitId)
		return nil, nil
	}

	lines := &Lines{}
	lines.Line = make([]string, 0)

	logfile.ReadLines(func(line string) (bool, error) {
		lines.Line = append(lines.Line, line)
		return true, nil
	})

	return lines, nil
}


func main() {
	juju.InitJujuHome()

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	NewRestEndpoint("/charm/", (*EndpointCharms)(nil))

	log.Fatal("Error serving HTTP", s.ListenAndServe())
}
