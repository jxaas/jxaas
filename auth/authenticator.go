package auth

import "net/http"

type Authenticator interface {
	Authenticate(tenantId string, req *http.Request) *Authentication
}
