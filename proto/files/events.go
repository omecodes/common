package filespb

type EventHandler interface {
	OnEvent(event *Event)
}

type handlerFunc struct {
	f func(*Event)
}

func (h *handlerFunc) OnEvent(event *Event) {
	h.f(event)
}

func EventHandlerFunc(f func(*Event)) EventHandler {
	return &handlerFunc{
		f: f,
	}
}

type Watcher interface {
	Watch(EventHandler) error
	Stop() error
}
