package juju

import (
	"fmt"

	"github.com/justinsb/gova/log"

	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/state/api"
)

var connectionError = `Unable to connect to environment "%s".
Please check your credentials or use 'juju bootstrap' to create a new environment.

Error details:
%v
`

func Init() error {
	return juju.InitJujuHome()
}

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

func (self *Client) ServiceDestroy(serviceId string) error {
	if !self.canAccess(serviceId) {
		return nil
	}

	return self.api.ServiceDestroy(serviceId)
}
