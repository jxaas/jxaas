package auth

import "net/http"

type MultiAuthenticator struct {
	Authenticator

	authenticators []Authenticator
}

func NewMultiAuthenticator(authenticators []Authenticator) *MultiAuthenticator {
	self := &MultiAuthenticator{}
	self.authenticators = authenticators
	return self
}

func (self *MultiAuthenticator) Authenticate(tenantId string, req *http.Request) *Authentication {
	for _, authenticator := range self.authenticators {
		authorization := authenticator.Authenticate(tenantId, req)
		if authorization != nil {
			return authorization
		}
	}
	return nil
}
