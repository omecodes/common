package ga

type CredentialsVerifyFunc func (cred *Credentials) (bool, error)

type Credentials struct {
	Username string
	Password string
}

type ProxyCredentials struct {
	Key    string
	Secret string
}
