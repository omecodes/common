package messaging

import (
	"github.com/omecodes/common/doers"
	"github.com/omecodes/common/log"
	pb "github.com/omecodes/common/messaging/proto"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func Serve(l net.Listener) (doers.Stopper, error) {
	h := &handler{
		store:         newMemMessageStore(),
		stopRequested: false,
	}

	server := grpc.NewServer()
	pb.RegisterNodesServer(server, h)

	err := server.Serve(l)
	if err != nil {
		log.Error("grpc::msg serve failed", err)
		return nil, err
	}

	return doers.StopFunc(h.Stop), nil
}

type handler struct {
	broadcastMutex     sync.Mutex
	keyCounter         int
	store              pb.Messages
	stopRequested      bool
	broadcastReceivers map[int]chan *pb.SyncMessage
	stoppers           []doers.Stopper
}

func (h *handler) Sync(stream pb.Nodes_SyncServer) error {
	broadcastReceiver := make(chan *pb.SyncMessage, 30)
	id := h.saveBroadcastReceiver(broadcastReceiver)
	defer h.deleteBroadcastReceiver(id)

	s := NewServerStreamSession(stream, broadcastReceiver, h.store, pb.MessageHandlerFunc(h.Handle))
	s.sync()
	return nil
}

func (h *handler) Handle(msg *pb.SyncMessage) {
	h.broadcastMutex.Lock()
	defer h.broadcastMutex.Unlock()
	for _, r := range h.broadcastReceivers {
		r <- msg
	}
}

func (h *handler) saveBroadcastReceiver(channel chan *pb.SyncMessage) int {
	h.broadcastMutex.Lock()
	defer h.broadcastMutex.Unlock()
	h.keyCounter++
	h.broadcastReceivers[h.keyCounter] = channel
	return h.keyCounter
}

func (h *handler) deleteBroadcastReceiver(key int) {
	h.broadcastMutex.Lock()
	defer h.broadcastMutex.Unlock()
	c := h.broadcastReceivers[key]
	defer close(c)
	delete(h.broadcastReceivers, key)
}

func (h *handler) Stop() error {
	for _, stopper := range h.stoppers {
		err := stopper.Stop()
		if err != nil {
			log.Error("msg::server stop failed", err)
		}
	}
	return nil
}

type memMessageStore struct {
	store *sync.Map
}

func (m *memMessageStore) Save(msg *pb.SyncMessage) error {
	m.store.Store(msg.Id, msg)
	return nil
}

func (m *memMessageStore) List() ([]*pb.SyncMessage, error) {
	var list []*pb.SyncMessage

	m.store.Range(func(key, value interface{}) bool {
		list = append(list, value.(*pb.SyncMessage))
		return true
	})

	return list, nil
}

func (m *memMessageStore) Invalidate(id string) error {
	m.store.Delete(id)
	return nil
}

func newMemMessageStore() *memMessageStore {
	s := new(memMessageStore)
	s.store = &sync.Map{}
	return s
}
