package messaging

import (
	"context"
	"crypto/tls"
	"github.com/google/uuid"
	"github.com/omecodes/common/codec"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/log"
	pb "github.com/omecodes/common/messaging/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"sync"
	"time"
)

type ClientStreamSession struct {
	syncMutex    sync.Mutex
	handlersLock sync.Mutex

	connectionAttempts int
	unconnectedTime    time.Time

	sendCloseSignal chan bool
	outboundStream  chan *pb.SyncMessage
	messageHandlers map[string]pb.MessageHandler

	syncing       bool
	stopRequested bool

	serverAddress string
	tlsConfig     *tls.Config
	conn          *grpc.ClientConn
	client        pb.NodesClient
}

func (r *ClientStreamSession) connect() error {
	if r.conn != nil && r.conn.GetState() == connectivity.Ready {
		return nil
	}

	var opts []grpc.DialOption
	if r.tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(r.tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	var err error
	r.conn, err = grpc.Dial(r.serverAddress, opts...)
	if err != nil {
		return err
	}
	r.client = pb.NewNodesClient(r.conn)
	return nil
}

func (r *ClientStreamSession) sync() {
	if r.isSyncing() {
		return
	}
	r.setSyncing()

	for !r.stopRequested {
		err := r.connect()
		if err != nil {
			time.After(time.Second * 2)
			continue
		}
		r.work()
	}
}

func (r *ClientStreamSession) work() {
	r.sendCloseSignal = make(chan bool)
	r.outboundStream = make(chan *pb.SyncMessage, 30)
	defer close(r.outboundStream)

	r.connectionAttempts++

	stream, err := r.client.Sync(context.Background())
	if err != nil {
		r.conn = nil
		if r.connectionAttempts == 1 {
			r.unconnectedTime = time.Now()
			log.Error("grpc::msg unconnected", errors.Errorf("%d", status.Code(err)))
			log.Info("grpc::msg trying again...")
		}
		return
	}
	defer stream.CloseSend()

	if r.connectionAttempts > 1 {
		log.Info("grpc::msg connected", log.Field("after", time.Since(r.unconnectedTime).String()), log.Field("attempts", r.connectionAttempts))
	} else {
		log.Info("grpc::msg connected")
	}
	r.connectionAttempts = 0

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go r.recv(stream, wg)
	go r.send(stream, wg)
	wg.Wait()
}

func (r *ClientStreamSession) send(stream pb.Nodes_SyncClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for !r.stopRequested {
		select {
		case <-r.sendCloseSignal:
			log.Info("grpc::msg stop send")
			return

		case event, open := <-r.outboundStream:
			if !open {
				return
			}

			err := stream.Send(event)
			if err != nil {
				if err != io.EOF {
					log.Error("grpc::msg send event", err)
				}
				return
			}
		}
	}
}

func (r *ClientStreamSession) recv(stream pb.Nodes_SyncClient, wg *sync.WaitGroup) {
	defer wg.Done()
	for !r.stopRequested {
		msg, err := stream.Recv()
		if err != nil {
			r.sendCloseSignal <- true
			close(r.sendCloseSignal)
			if err != io.EOF {
				log.Error("grpc::msg recv event", err)
			}
			return
		}

		for _, h := range r.messageHandlers {
			go h.Handle(msg)
		}

		log.Info("grpc::msg new event", log.Field("type", msg.Type), log.Field("id", msg.Id))
	}
}

func (r *ClientStreamSession) isSyncing() bool {
	r.syncMutex.Lock()
	defer r.syncMutex.Unlock()
	return r.syncing
}

func (r *ClientStreamSession) setSyncing() {
	r.syncMutex.Lock()
	defer r.syncMutex.Unlock()
	r.syncing = true
}

func (r *ClientStreamSession) RegisterHandler(h pb.MessageHandler) string {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()
	hid := uuid.New().String()
	r.messageHandlers[hid] = h
	return hid
}

func (r *ClientStreamSession) DeRegisterHandler(id string) {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()
	delete(r.messageHandlers, id)
}

func (r *ClientStreamSession) Send(msgType string, name string, o interface{}) error {
	if !r.isSyncing() {
		return errors.New("not connected")
	}

	encoded, err := codec.Json.Encode(o)
	if err != nil {
		return err
	}
	r.outboundStream <- &pb.SyncMessage{
		Type:    msgType,
		Id:      name,
		Encoded: encoded,
	}
	return nil
}

func (r *ClientStreamSession) SendMsg(msg *pb.SyncMessage) error {
	if !r.isSyncing() {
		return errors.New("not connected")
	}
	r.outboundStream <- msg
	return nil
}

func NewClientStreamSession(address string, tlsConfig *tls.Config) *ClientStreamSession {
	sc := &ClientStreamSession{
		tlsConfig:       tlsConfig,
		serverAddress:   address,
		messageHandlers: map[string]pb.MessageHandler{},
	}
	go sc.sync()
	return sc
}
