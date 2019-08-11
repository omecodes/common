package app

type Node interface {
	Configure(args *ConfigVars) error
	Init(args *Vars) error
	Start() error
	Stop()
}
