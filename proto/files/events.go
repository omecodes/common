package filespb

type EventHandler interface {
	OnEvent(event *Event) error
}

type handlerFunc struct {
	f func(*Event) error
}

func (h *handlerFunc) OnEvent(event *Event) error {
	return h.f(event)
}

func EventHandlerFunc(f func(*Event) error) EventHandler {
	return &handlerFunc{
		f: f,
	}
}

type Watcher interface {
	Watch(EventHandler) error
	Stop() error
}
