package endpoints

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"
	"launchpad.net/goose/identity"
)

type EndpointXaas struct {
}

func (self *EndpointXaas) Item(key string, req *http.Request) (*EndpointTenant, error) {
	child := &EndpointTenant{}

	tenantName := strings.Replace(key, "-", "", -1)

	var authorization *identity.AuthDetails

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

				// XXX: Don't hard-code
				authServerUrl := "http://127.0.0.1:5000/v2.0"

				authenticator := identity.NewAuthenticator(identity.AuthUserPass, nil)
				creds := identity.Credentials{TenantName: tenantName, User: username, URL: authServerUrl + "/tokens", Secrets: password}
				auth, err := authenticator.Auth(&creds)
				if err != nil {
					log.Warn("Error authenticating against Openstack Identity", err)
				} else if auth == nil {
					log.Warn("Auth returned from Openstack identity was nil")
				} else {
					log.Debug("Got auth token: %v", auth.TenantId)
					authorization = auth
				}
			}
		}
	}

	if authorization == nil {
		notAuthorized := rs.HttpError(http.StatusUnauthorized)
		notAuthorized.Headers["WWW-Authenticate"] = "Basic realm=\"jxaas\""
		return nil, notAuthorized
	}

	child.Tenant = tenantName
	// TODO: Use tenantId? authorization.TenantId

	return child, nil
}
