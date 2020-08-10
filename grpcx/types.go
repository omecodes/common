package grpcx

type CredentialsVerifyFunc func(cred *Credentials) (bool, error)

type ProxyCredentialsVerifyFunc func(cred *ProxyCredentials) (bool, error)

type Credentials struct {
	Username string
	Password string
}

type ProxyCredentials struct {
	Key    string
	Secret string
}
