package auth

func NewOpenstackMultiAuthenticator(keystoneEndpoint string) Authenticator {
	keystoneEndpoint = toKeystoneEndpoint(keystoneEndpoint)

	basic := NewOpenstackBasicAuthenticator(keystoneEndpoint)
	token := NewOpenstackTokenAuthenticator(keystoneEndpoint)
	return NewMultiAuthenticator([]Authenticator{basic, token})
}
