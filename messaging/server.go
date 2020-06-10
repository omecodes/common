package messaging

import (
	"github.com/omecodes/common/doers"
	"github.com/omecodes/common/log"
	pb "github.com/omecodes/common/messaging/proto"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func Serve(l net.Listener, handler pb.MessageHandler) (doers.Stopper, error) {
	h := &server{
		messages:           newMemMessageStore(),
		stopRequested:      false,
		messageHandler:     handler,
		broadcastReceivers: map[int]chan *pb.SyncMessage{},
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

type server struct {
	broadcastMutex     sync.Mutex
	keyCounter         int
	messages           pb.Messages
	stopRequested      bool
	stoppers           []doers.Stopper
	broadcastReceivers map[int]chan *pb.SyncMessage
	messageHandler     pb.MessageHandler
}

func (s *server) Sync(stream pb.Nodes_SyncServer) error {
	broadcastReceiver := make(chan *pb.SyncMessage, 30)
	id := s.saveBroadcastReceiver(broadcastReceiver)
	defer s.deleteBroadcastReceiver(id)

	sess := NewServerStreamSession(stream, broadcastReceiver, s.messages, pb.MessageHandlerFunc(s.Handle))
	sess.sync()
	return nil
}

func (s *server) Handle(msg *pb.SyncMessage) {
	s.broadcastMutex.Lock()
	defer s.broadcastMutex.Unlock()
	for _, r := range s.broadcastReceivers {
		r <- msg
	}
}

func (s *server) saveBroadcastReceiver(channel chan *pb.SyncMessage) int {
	s.broadcastMutex.Lock()
	defer s.broadcastMutex.Unlock()
	s.keyCounter++
	s.broadcastReceivers[s.keyCounter] = channel
	return s.keyCounter
}

func (s *server) deleteBroadcastReceiver(key int) {
	s.broadcastMutex.Lock()
	defer s.broadcastMutex.Unlock()
	c := s.broadcastReceivers[key]
	defer close(c)
	delete(s.broadcastReceivers, key)
}

func (s *server) broadcast(msg *pb.SyncMessage) {

}

func (s *server) Stop() error {
	for _, stopper := range s.stoppers {
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
