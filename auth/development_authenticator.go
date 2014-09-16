package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type DevelopmentAuthenticator struct {
	Authenticator
}

func NewDevelopmentAuthenticator() *DevelopmentAuthenticator {
	self := &DevelopmentAuthenticator{}
	return self
}

func (self *DevelopmentAuthenticator) Authenticate(tenantId string, req *http.Request) *Authentication {
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

				if username == password {
					authorization = &Authentication{TenantId: tenantId, TenantName: tenantId}
				}
			}
		}
	}

	return authorization
}
