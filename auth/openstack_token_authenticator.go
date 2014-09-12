package auth

import (
	"net/http"
	"strings"

	"github.com/justinsb/gova/log"
	"launchpad.net/goose/identity"
)

type OpenstackTokenAuthenticator struct {
	Authenticator

	keystoneEndpoint string
}

func NewOpenstackTokenAuthenticator(keystoneEndpoint string) *OpenstackTokenAuthenticator {
	keystoneEndpoint = toKeystoneEndpoint(keystoneEndpoint)

	self := &OpenstackTokenAuthenticator{}
	self.keystoneEndpoint = keystoneEndpoint
	return self
}

func (self *OpenstackTokenAuthenticator) Authenticate(tenantId string, req *http.Request) *Authentication {
	var authorization *Authentication

	authTokens := req.Header["X-Auth-Token"]
	if len(authTokens) > 0 {
		authToken := strings.TrimSpace(authTokens[0])

		log.Debug("Request to authenticate with token: %v in tenant: %v", authToken, tenantId)

		tenants, err := identity.ListTenantsForToken(self.keystoneEndpoint+"tenants", authToken, nil)
		if err != nil {
			log.Warn("Error authenticating against Openstack Identity", err)
		} else if tenants == nil {
			log.Warn("Tenants returned from Openstack identity was nil")
		} else {
			for _, tenant := range tenants {
				if tenant.Id == tenantId {
					if !tenant.Enabled {
						log.Warn("In project, but not enabled for project: %v", tenantId)
						continue
					}
					authorization = &Authentication{TenantId: tenant.Id, TenantName: tenant.Name}
					break
				}
			}

			if authorization == nil {
				log.Warn("Valid token, but not authorized for project: %v", tenantId)
			}
		}
	}
	return authorization
}
