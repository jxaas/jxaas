package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/justinsb/gova/log"
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

func (self *OpenstackBasicAuthenticator) Authenticate(tenantId string, req *http.Request) *Authentication {
	var authorization *Authentication

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

				log.Debug("Request to authenticate as: %v in tenant %v", username, tenantId)

				authenticator := identity.NewAuthenticator(identity.AuthUserPass, nil)
				creds := identity.Credentials{TenantId: tenantId, User: username, URL: self.keystoneEndpoint + "tokens", Secrets: password}
				auth, err := authenticator.Auth(&creds)
				if err != nil {
					log.Warn("Error authenticating against Openstack Identity", err)
				} else if auth == nil {
					log.Warn("Auth returned from Openstack identity was nil")
				} else {
					log.Debug("Got auth token: %v", auth.TenantId)
					authorization = &Authentication{TenantId: auth.TenantId, TenantName: auth.TenantName}

				}
			}
		}
	}

	return authorization
}
