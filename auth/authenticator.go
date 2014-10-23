package auth

import "net/http"

type Authenticator interface {
	// Authenticates the request
	// tenantSpec is the tenant portion of the url.
	// The meaning is actually Authenticator dependent:
	//  it is an ID for OpenStack token based auth (because we use the OpenStack tenant id)
	//  it is a name for OpenStack basic auth (because the user probably doesn't know their tenant id)
	//  it is an ID for development auth (because we have no way to map it)
	Authenticate(tenantSpec string, req *http.Request) *Authentication
}
