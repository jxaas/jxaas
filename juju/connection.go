package juju

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/constraints"
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
	api *api.Client
}

func ClientFactory() (*Client, error) {
	envName := cmd.ReadCurrentEnvironment()
	apiclient, err := juju.NewAPIClientFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}

	wrapper := &Client{}
	wrapper.api = apiclient
	//defer apiclient.Close()
	return wrapper, err
}

func (self *Client) canAccess(serviceId string) bool {
	log.Warn("Juju connection canAccess is stub-implemented")
	return true
}

func (self *Client) GetStatus(serviceId string) (*api.ServiceStatus, error) {
	if !self.canAccess(serviceId) {
		return nil, nil
	}

	// TODO: Is this efficient?  Any direct just-this-service call?
	patterns := make([]string, 1)
	patterns[0] = serviceId
	status, err := self.api.Status(patterns)

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

func (self *Client) FindConfig(serviceId string) (*params.ServiceGetResults, error) {
	if !self.canAccess(serviceId) {
		return nil, nil
	}

	config, err := self.api.ServiceGet(serviceId)
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

	err := self.api.ServiceSet(serviceId, options)
	if err != nil {
		return err
	}

	return nil
}

func (self *Client) ListServices() (*api.Status, error) {
	patterns := []string{}
	status, err := self.api.Status(patterns)

	if err != nil {
		return nil, err
	}

	// TODO: Filter to just our env
	log.Warn("Service filtering not implemented")

	return status, nil
}

func (self *Client) ServiceDestroy(serviceId string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	return self.api.ServiceDestroy(serviceId)
}

func (self *Client) ServiceDeploy(charmUrl string, serviceId string, numUnits int, configYAML string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	var constraints constraints.Value
	var toMachineSpec string

	return self.api.ServiceDeploy(charmUrl, serviceId, numUnits, configYAML, constraints, toMachineSpec)

	//	if params.IsCodeNotImplemented(err) {
	//		logger.Infof("Status not supported by the API server, " +
	//			"falling back to 1.16 compatibility mode " +
	//			"(direct DB access)")
	//		status, err = c.getStatus1dot16()
	//	}

}
