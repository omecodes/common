package app

type Node interface {
	Configure(args *ConfigArgs) error
	Init(args *RunArgs) error
	Start() error
	Stop()
}
