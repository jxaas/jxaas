package jxaas

import "github.com/jxaas/jxaas/juju"

// A JXaaS instance
type Instance interface {
	GetJujuClient() *juju.Client
}
