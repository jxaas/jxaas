package juju

import (
	"fmt"
	"strings"

	"github.com/justinsb/gova/log"

	"bitbucket.org/jsantabarbara/jxaas/model"

	"launchpad.net/juju-core/cmd/envcmd"
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

const (
	PREFIX_RELATIONINFO = "__jxaas_relinfo_"
)

func Init() error {
	return juju.InitJujuHome()
}

// Client is a simple wrapper around the Juju API.
// It is responsible for enforcing multi-tenancy security,
// and other additional concerns we have.
type Client struct {
	state  *api.State
	client *api.Client
}

func ClientFactory() (*Client, error) {
	envName := envcmd.ReadCurrentEnvironment()

	state, err := juju.NewAPIFromName(envName)
	if err != nil {
		return nil, fmt.Errorf(connectionError, envName, err)
	}

	client := state.Client()

	wrapper := &Client{}
	wrapper.client = client
	wrapper.state = state
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

func (self *Client) ListServices() (*api.Status, error) {
	patterns := []string{}
	status, err := self.client.Status(patterns)

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

	return self.client.ServiceDestroy(serviceId)
}

func (c *Client) call(method string, params, result interface{}) error {
	return c.state.Call("Client", "", method, params, result)
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
	return self.client.CharmInfo(charmUrl)
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
		return nil, err
	}

	return results, nil
}

func (self *Client) SetRelationInfo(serviceId, unitId, relationId string, properties map[string]string) error {
	if !self.canAccess(serviceId) {
		return fmt.Errorf("Service not found")
	}

	// Annotations on relations aren't supported, and it is tricky to get the relation id
	// So tag it on the service instead
	annotateTag := "service-" + serviceId

	pairs := make(map[string]string)
	for k, v := range properties {
		pairs[PREFIX_RELATIONINFO+unitId+"_"+relationId+"_"+k] = v
	}

	log.Info("Setting annotations on %v: %v", annotateTag, pairs)

	err := self.client.SetAnnotations(annotateTag, pairs)
	if err != nil {
		log.Warn("Error setting annotations", err)
		// TODO: Mask error?
		return err
	}

	return nil
}

func (self *Client) GetRelationInfo(serviceId string, relationKey string) (*model.RelationInfo, error) {
	if !self.canAccess(serviceId) {
		return nil, fmt.Errorf("Service not found")
	}

	annotateTag := "service-" + serviceId

	annotations, err := self.client.GetAnnotations(annotateTag)
	if err != nil {
		log.Warn("Error getting annotations", err)
		// TODO: Mask error?
		return nil, err
	}

	relationIdPrefix := relationKey + ":"

	relationInfo := &model.RelationInfo{}
	relationInfo.Properties = make(map[string]string)

	for tagName, v := range annotations {
		if !strings.HasPrefix(tagName, PREFIX_RELATIONINFO) {
			//log.Debug("Prefix mismatch: %v", tagName)
			continue
		}
		suffix := tagName[len(PREFIX_RELATIONINFO):]
		tokens := strings.SplitN(suffix, "_", 3)
		if len(tokens) < 3 {
			log.Debug("Ignoring unparseable tag: %v", tagName)
			continue
		}

		// unitId = tokens[0]
		relationId := tokens[1]
		if !strings.HasPrefix(relationId, relationIdPrefix) {
			//log.Debug("Relation prefix mismatch: %v", relationId)
			continue
		}

		key := tokens[2]
		relationInfo.Properties[key] = v
	}

	return relationInfo, nil
}
