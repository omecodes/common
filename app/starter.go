package app

type starterFunc struct {
	f func() error
}

func (sf *starterFunc) Start() error {
	return sf.f()
}

type Starter interface {
	Start() error
}

func NewStarter(sf func() error) Starter {
	return &starterFunc{f: sf}
}
