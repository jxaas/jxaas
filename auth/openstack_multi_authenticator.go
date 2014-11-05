package auth

func NewOpenstackMultiAuthenticator(keystoneUrl string) *MultiAuthenticator {
	token := NewOpenstackTokenAuthenticator(keystoneUrl)
	basic := NewOpenstackBasicAuthenticator(keystoneUrl)
	return NewMultiAuthenticator([]Authenticator{token, basic})
}
