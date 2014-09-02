package endpoints

import (
	"encoding/base64"
	"net/http"
	"strings"
	"github.com/justinsb/gova/rs"
)

type EndpointXaas struct {
}

func (self *EndpointXaas) Item(key string, req *http.Request) (*EndpointTenant, error) {
	child := &EndpointTenant{}

	var authorization *string

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
					authorization = &username
				}
			}
		}
	}

	tenant := strings.Replace(key, "-", "", -1)

	if authorization == nil || *authorization != tenant {
		notAuthorized := rs.HttpError(http.StatusUnauthorized)
		notAuthorized.Headers["WWW-Authenticate"] = "Basic realm=\"jxaas\""
		return nil, notAuthorized
	}

	child.Tenant = tenant

	return child, nil
}
