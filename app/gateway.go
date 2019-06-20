package app

type Gateway interface {
	Name() string
	Protocol() string
	Port() int
	Start() error
	Stop()
}
