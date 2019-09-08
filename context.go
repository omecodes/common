package common

type contextKey string

const (
	ContextToken     = contextKey("token")
	ContextUserAgent = contextKey("user-agent")
)
