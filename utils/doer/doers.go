package doer

type StopFunc func() error

func (sf StopFunc) Stop() error {
	return sf()
}

type Stopper interface {
	Stop() error
}

type CleanStopper interface {
	Stop()
}
