package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/justinsb/gova/log"
	"launchpad.net/goose/errors"
	"launchpad.net/goose/identity"
)

type OpenstackBasicAuthenticator struct {
	Authenticator

	keystoneEndpoint string
}

func toKeystoneEndpoint(keystoneEndpoint string) string {
	// TODO: Should we add /v2.0 ?
	// e.g. "http://127.0.0.1:5000/v2.0"
	if !strings.HasSuffix(keystoneEndpoint, "/") {
		keystoneEndpoint += "/"
	}
	return keystoneEndpoint
}

func NewOpenstackBasicAuthenticator(keystoneEndpoint string) *OpenstackBasicAuthenticator {
	keystoneEndpoint = toKeystoneEndpoint(keystoneEndpoint)

	self := &OpenstackBasicAuthenticator{}
	self.keystoneEndpoint = keystoneEndpoint
	return self
}

// Authenticate against Openstack using basic auth
func (self *OpenstackBasicAuthenticator) Authenticate(tenantSpec string, req *http.Request) *Authentication {
	var authorization *Authentication

	// Because the user hasn't authenticated with keystone, we assume tenantSpec is actually a tenant _name_ here,
	// not a tenant id
	tenantName := tenantSpec

	authorizationHeaders := req.Header["Authorization"]
	if len(authorizationHeaders) > 0 {
		authorizationHeader := strings.TrimSpace(authorizationHeaders[0])

		tokens := strings.SplitN(authorizationHeader, " ", 2)
		if len(tokens) == 2 && tokens[0] == "Basic" {
			payload, _ := base64.StdEncoding.DecodeString(tokens[1])
			usernameAndPassword := strings.SplitN(string(payload), ":", 2)

			if len(usernameAndPassword) == 2 {
				username := usernameAndPassword[0]
				password := usernameAndPassword[1]

				log.Debug("Request to authenticate as: %v in tenant %v", username, tenantName)

				authenticator := identity.NewAuthenticator(identity.AuthUserPass, nil)
				creds := identity.Credentials{TenantName: tenantName, User: username, URL: self.keystoneEndpoint + "tokens", Secrets: password}
				auth, err := authenticator.Auth(&creds)
				if err != nil {
					if errors.IsUnauthorised(err) {
						log.Debug("Openstack Identity rejected the authentication request (401)")
					} else {
						log.Warn("Error authenticating against Openstack Identity", err)
					}
				} else if auth == nil {
					log.Warn("Auth returned from Openstack identity was nil")
				} else {
					if auth.TenantId != "" {
						authorization = &Authentication{TenantId: auth.TenantId, TenantName: auth.TenantName}
					}

					// We don't _need_ to use TenantName based auth; we could retrieve the tenants like this...
					//					tenants, err := identity.ListTenantsForToken(self.keystoneEndpoint+"tenants", auth.Token, nil)
					//					if err != nil {
					//						log.Warn("Unable to fetch tenants for token", err)
					//					} else {
					//						log.Debug("Got tenants: %v", tenants)
					//						for _, tenant := range tenants {
					//							if tenant.Name == tenantName {
					//								authorization = &Authentication{TenantId: tenant.Id, TenantName: tenant.Name}
					//								break
					//							}
					//						}
					//						if authorization == nil {
					//							log.Debug("Authenticated with keystone, but not authorized for project: %v", tenantName)
					//						}
					//					}
				}
			}
		}
	}

	return authorization
}
