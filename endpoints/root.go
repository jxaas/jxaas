package endpoints

import (
	"net/http"
	"strings"

	"github.com/justinsb/gova/assert"
	"github.com/justinsb/gova/inject"
	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/rs"
	"github.com/jxaas/jxaas/auth"
)

type EndpointXaas struct {
	Authenticator auth.Authenticator `inject:"y"`
}

type Authorization struct {
	TenantId   string
	TenantName string
}

func (self *EndpointXaas) Item(key string, injector inject.Injector, req *http.Request) (*EndpointTenant, error) {
	child := &EndpointTenant{}

	tenantId := key
	tenantName := strings.Replace(key, "-", "", -1)

	// TODO: Implement authz
	assert.That(self.Authenticator != nil)
	authentication := self.Authenticator.Authenticate(tenantId, req)

	if authentication == nil {
		log.Debug("Authentication failed")
		notAuthorized := rs.HttpError(http.StatusUnauthorized)
		notAuthorized.Headers["WWW-Authenticate"] = "Basic realm=\"jxaas\""
		return nil, notAuthorized
	} else {
		child.Tenant = tenantName
		// TODO: Use tenantId? authorization.TenantId

		return child, nil
	}
}
