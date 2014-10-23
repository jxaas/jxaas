package juju

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/justinsb/gova/files"
	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/cmd/envcmd"
	"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/api/params"
)

var connectionError = `Unable to connect to environment "%s".
Please check your credentials or use 'juju bootstrap' to create a new environment.

Error details:
%v
`

func Init() error {
	return juju.InitJujuHome()
}

// Client is a simple wrapper around the Juju API.
// It is responsible for enforcing multi-tenancy security,
// and other additional concerns we have.
type Client struct {
	apiState *api.State
	client   *api.Client
}

func SimpleClientFactory(info *api.Info) (*Client, error) {
	dialOpts := api.DialOpts{}
	state, err := api.Open(info, dialOpts)
	if err != nil {
		return nil, err
	}

	client := state.Client()

	wrapper := &Client{}
	wrapper.client = client
	wrapper.apiState = state
	//defer apiclient.Close()
	return wrapper, err
}

func EnvClientFactory() (*Client, error) {
	envName := envcmd.ReadCurrentEnvironment()

	state, err := juju.NewAPIFromName(envName)
	if err != nil {
		log.Warn("Got error building API from name: %v", envName, err)
		return nil, fmt.Errorf(connectionError, envName, err)
	}

	client := state.Client()

	wrapper := &Client{}
	wrapper.client = client
	wrapper.apiState = state
	//defer apiclient.Close()
	return wrapper, err
}

func DirectClientFactory(conf *config.Config) (*Client, error) {
	env, err := environs.New(conf)
	if err != nil {
		return nil, err
	}

	dialOpts := api.DefaultDialOpts()
	conn, err := juju.NewAPIConn(env, dialOpts)
	if err != nil {
		return nil, err
	}

	wrapper := &Client{}
	wrapper.client = conn.State.Client()
	wrapper.apiState = conn.State
	//defer apiclient.Close()
	return wrapper, err
}

func (self *Client) canAccess(serviceId string) bool {
	// Maybe we should panic here
	log.Warn("Juju connection canAccess is stub-implemented")
	return true
}

func (self *Client) canAccessPrefix(serviceId string) bool {
	// Maybe we should panic here
	log.Warn("Juju connection canAccessPrefix is stub-implemented")
	return true
}

func (self *Client) GetSystemStatus() (*api.Status, error) {
	patterns := make([]string, 0)
	status, err := self.client.Status(patterns)

	if err != nil {
		return nil, err
	}

	return status, nil
}

func (self *Client) DestroyMachine(machineId string) error {
	err := self.client.DestroyMachines(machineId)

	if err != nil {
		return err
	}
	return nil
}

func (self *Client) GetServiceStatus(serviceId string) (*api.ServiceStatus, error) {
	if !self.canAccess(serviceId) {
		return nil, nil
	}

	// TODO: Is this efficient?  Any direct just-this-service call?
	patterns := make([]string, 1)
	patterns[0] = serviceId
	status, err := self.client.Status(patterns)

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

	state, found := status.Services[serviceId]
	if !found {
		return nil, nil
	}

	return &state, nil
}

func (self *Client) GetServiceStatusList(prefix string) (map[string]api.ServiceStatus, error) {
	if !self.canAccessPrefix(prefix) {
		return nil, nil
	}

	patterns := make([]string, 1)
	patterns[0] = prefix + "*"
	status, err := self.client.Status(patterns)

	// Display any error, but continue to print status if some was returned
	if err != nil {
		return nil, err
	}

	return status.Services, nil
}

func (self *Client) FindConfig(serviceId string) (*params.ServiceGetResults, error) {
	if !self.canAccess(serviceId) {
		return nil, nil
	}

	config, err := self.client.ServiceGet(serviceId)
	if err != nil {
		paramsError, ok := err.(*params.Error)
		if ok && paramsError.Code == "not found" {
			// Treat as not-an-error
			return nil, nil
		}
		return nil, err
	}

	return config, nil
}

func (self *Client) SetConfig(serviceId string, options map[string]string) error {
	if !self.canAccess(serviceId) {
		return fmt.Errorf("Unknown service: %v", serviceId)
	}

	err := self.client.ServiceSet(serviceId, options)
	if err != nil {
		return err
	}

	return nil
}

func (self *Client) SetExposed(serviceId string, exposed bool) error {
	if !self.canAccess(serviceId) {
		return fmt.Errorf("Unknown service: %v", serviceId)
	}

	var err error
	if exposed {
		err = self.client.ServiceExpose(serviceId)
	} else {
		err = self.client.ServiceUnexpose(serviceId)
	}

	if err != nil {
		return err
	}

	return nil
}

func (self *Client) ServiceDestroy(serviceId string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	return self.client.ServiceDestroy(serviceId)
}

func (c *Client) call(method string, params, result interface{}) error {
	return c.apiState.Call("Client", "", method, params, result)
}

// Fixed so that we can omit numUnits (by passing -1)
func (c *Client) serviceDeploy(charmURL string, serviceName string, numUnits int, configYAML string, cons constraints.Value, toMachineSpec string) error {
	params := params.ServiceDeploy{
		ServiceName:   serviceName,
		CharmUrl:      charmURL,
		ConfigYAML:    configYAML,
		Constraints:   cons,
		ToMachineSpec: toMachineSpec,
	}
	if numUnits >= 0 {
		params.NumUnits = numUnits
	}

	return c.call("ServiceDeploy", params, nil)
}

func (self *Client) ServiceDeploy(charmUrl string, serviceId string, numUnits int, configYAML string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	var constraints constraints.Value
	var toMachineSpec string

	return self.serviceDeploy(charmUrl, serviceId, numUnits, configYAML, constraints, toMachineSpec)

	//	if params.IsCodeNotImplemented(err) {
	//		logger.Infof("Status not supported by the API server, " +
	//			"falling back to 1.16 compatibility mode " +
	//			"(direct DB access)")
	//		status, err = c.getStatus1dot16()
	//	}
}

func (self *Client) CharmInfo(charmUrl string) (*api.CharmInfo, error) {
	// TODO: Caching?
	defaultSeries := "precise"
	localRepoPath := ""
	return getCharmInfo(self.client, charmUrl, localRepoPath, defaultSeries)
	//return self.client.CharmInfo(charmUrl)
}

func (self *Client) PutRelation(from, to string) (*params.AddRelationResults, error) {
	if !self.canAccess(from) {
		return nil, fmt.Errorf("Cannot find service")
	}

	if !self.canAccess(to) {
		return nil, fmt.Errorf("Cannot find service")
	}

	results, err := self.client.AddRelation(from, to)

	if err != nil {
		jujuError, ok := err.(*params.Error)
		if ok {
			// There is no code :-(
			//			if jujuError.Code == "relation already exists" {
			//				return nil, nil
			//			}
			if strings.HasSuffix(jujuError.Message, "relation already exists") {
				return nil, nil
			}
			log.Debug("Error while creating relation from %v to %v: Code=%v Message=%v", from, to, jujuError.Code, jujuError.Message)
		}
		return nil, err
	}

	return results, nil
}

// Adds annotations on the specified service
func (self *Client) SetServiceAnnotations(serviceId string, pairs map[string]string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	annotateTag := "service-" + serviceId

	return self.client.SetAnnotations(annotateTag, pairs)
}

// Deletes annotations from the specified service
func (self *Client) DeleteServiceAnnotations(serviceId string, keys []string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	annotateTag := "service-" + serviceId

	pairs := map[string]string{}
	for _, key := range keys {
		pairs[key] = ""
	}

	return self.client.SetAnnotations(annotateTag, pairs)
}

// Retrieves all annotations on the service
func (self *Client) GetServiceAnnotations(serviceId string) (map[string]string, error) {
	if !self.canAccess(serviceId) {
		return nil, nil
	}

	annotateTag := "service-" + serviceId

	annotations, err := self.client.GetAnnotations(annotateTag)
	return annotations, err
}

func (self *Client) AddServiceUnits(serviceId string, numUnits int) ([]string, error) {
	if !self.canAccess(serviceId) {
		return nil, fmt.Errorf("Unknown service: %v", serviceId)
	}

	machineSpecString := ""
	units, err := self.client.AddServiceUnits(serviceId, numUnits, machineSpecString)
	if err != nil {
		return nil, err
	}

	return units, nil
}

func (self *Client) DestroyUnit(serviceId string, unitId int) error {
	if !self.canAccess(serviceId) {
		return fmt.Errorf("Unknown service: %v", serviceId)
	}

	unitName := serviceId + "/" + strconv.Itoa(unitId)
	err := self.client.DestroyServiceUnits(unitName)
	if err != nil {
		return err
	}

	return nil
}

func (self *Client) Run(serviceId string, unitIds []string, command string, timeout time.Duration) ([]params.RunResult, error) {
	if !self.canAccess(serviceId) {
		return nil, fmt.Errorf("Unknown service: %v", serviceId)
	}

	params := params.RunParams{
		Commands: command,
		Timeout:  5 * time.Second,
		Machines: nil,
		Services: nil,
		Units:    nil,
	}

	if unitIds == nil {
		params.Services = []string{serviceId}
	} else {
		params.Units = []string{}
		for _, unitId := range unitIds {
			params.Units = append(params.Units, serviceId+"/"+unitId)
		}
	}

	results, err := self.client.Run(params)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func ParseUnit(unitId string) string {
	slash := strings.Index(unitId, "/")
	return unitId[slash+1:]
}

func (self *Client) GetLogStore() (*JujuLogStore, error) {
	// TODO: Cache?
	// TODO: SSH?

	baseDir := "/var/log/juju/"
	exists, err := files.Exists(baseDir)
	if err != nil {
		log.Warn("Error checking if /var/log/juju exists", err)
		return nil, err
	}

	logStore := &JujuLogStore{}

	if exists {
		logStore.BaseDir = baseDir
		return logStore, nil
	}

	// LXC looks like /var/log/juju-<username>-local/

	dirs, err := ioutil.ReadDir("/var/log")
	if err != nil {
		log.Warn("Error listing contents of /var/log", err)
		return nil, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		name := dir.Name()
		if strings.HasPrefix(name, "juju-") && strings.HasSuffix(name, "-local") {
			logStore.BaseDir = filepath.Join("/var/log", name)
			return logStore, nil
		}
	}

	return nil, errors.New("Unable to find juju log store")
}
