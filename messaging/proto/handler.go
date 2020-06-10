package pb

type MessageHandlerFunc func(*SyncMessage)

func (h MessageHandlerFunc) Handle(msg *SyncMessage) {
	h(msg)
}

type MessageHandler interface {
	Handle(*SyncMessage)
}

type Messages interface {
	Save(*SyncMessage) error
	List() ([]*SyncMessage, error)
	Invalidate(string) error
}
