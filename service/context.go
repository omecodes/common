package service

type ContextKey string

const (
	User            = ContextKey("user")
	Caller          = ContextKey("caller")
	Token           = ContextKey("token")
	PeerIp          = ContextKey("peer-ip")
	PeerCertificate = ContextKey("peer-certificate")
)
