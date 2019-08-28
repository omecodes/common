package discovery

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"github.com/zoenion/common/clone"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"log"
	"sync"
	"time"
)

type RegistryEventHandler interface {
	Handle(*pb.Event)
}

type eventHandlerFunc struct {
	f func(event *pb.Event)
}

func (hf *eventHandlerFunc) Handle(event *pb.Event) {
	hf.f(event)
}

func EventHandlerFunc(f func(*pb.Event)) RegistryEventHandler {
	return &eventHandlerFunc{f: f}
}

type SyncedRegistry struct {
	servicesLock sync.Mutex
	handlersLock sync.Mutex
	services     map[string]*pb.Info
	client       pb.RegistryClient

	tlsConfig     *tls.Config
	serverAddress string
	stop          bool
	conn          *grpc.ClientConn
	eventHandlers map[string]pb.RegistryEventHandler
}

func (r *SyncedRegistry) Connect() error {
	if r.conn != nil && r.conn.GetState() == connectivity.Ready {
		return nil
	}

	var opts []grpc.DialOption

	if r.tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(r.tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	attempt := 0
	for !r.stop || r.conn != nil && r.conn.GetState() != connectivity.Ready {
		attempt++
		var err error
		r.conn, err = grpc.Dial(r.serverAddress, opts...)
		if err != nil {
			log.Printf("connection to registry server failed: %s\n", err)
			<-time.After(time.Second * 3)
			if attempt == 3 {
				return fmt.Errorf("could not connect to server: %s", err)
			}
		} else {
			break
		}
	}

	r.client = pb.NewRegistryClient(r.conn)
	go r.connected()
	return nil
}

func (r *SyncedRegistry) Disconnect() error {
	r.stop = true
	r.disconnected()
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *SyncedRegistry) Register(i *pb.Info) (string, error) {
	err := r.Connect()
	if err != nil {
		return "", fmt.Errorf("could not connect to server: %s", err)
	}
	rsp, err := r.client.Register(context.Background(), &pb.RegisterRequest{Service: i})
	if err != nil {
		return "", err
	}
	return rsp.RegistryId, nil
}

func (r *SyncedRegistry) Deregister(id string) error {
	err := r.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to server: %s", err)
	}

	_, err = r.client.Deregister(context.Background(), &pb.DeregisterRequest{RegistryId: id})
	return err
}

func (r *SyncedRegistry) Get(id string) (*pb.Info, error) {
	return r.get(id), nil
}

func (r *SyncedRegistry) Certificate(id string) ([]byte, error) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	for _, s := range r.services {
		if id == fmt.Sprintf("%s.%s", s.Namespace, s.Name) {
			strCert, found := s.Meta["certificate"]
			if !found {
				return nil, errors.NotFound
			}
			return []byte(strCert), nil
		}
	}
	return nil, errors.NotFound
}

func (r *SyncedRegistry) ConnectionInfo(id string, protocol pb.Protocol) (*pb.ConnectionInfo, error) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	ci := new(pb.ConnectionInfo)

	for _, s := range r.services {
		if id == fmt.Sprintf("%s.%s", s.Namespace, s.Name) {
			for _, n := range s.Nodes {
				if n.Protocol == protocol {
					ci.Address = n.Address
					strCert, found := s.Meta["certificate"]
					if !found {
						return ci, nil
					}

					ci.Certificate = []byte(strCert)
					return ci, nil
				}
			}
		}
	}
	return nil, errors.NotFound
}

func (r *SyncedRegistry) RegisterEventHandler(h pb.RegistryEventHandler) string {
	r.handlersLock.Lock()
	defer r.handlersLock.Lock()
	hid := uuid.New().String()
	r.eventHandlers[hid] = h
	return hid
}

func (r *SyncedRegistry) DeregisterEventHandler(hid string) {
	r.handlersLock.Lock()
	defer r.handlersLock.Lock()
	delete(r.eventHandlers, hid)
}

func (r *SyncedRegistry) GetOfType(t pb.Type) ([]*pb.Info, error) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	var result []*pb.Info
	for _, s := range r.services {
		if s.Type == t {
			c := clone.New(s)
			result = append(result, c.(*pb.Info))
		}
	}
	return result, nil
}

func (r *SyncedRegistry) publishEvent(e pb.Event) {
	r.handlersLock.Lock()
	r.handlersLock.Unlock()

	for _, handler := range r.eventHandlers {
		handler.Handle(&e)
	}
}

func (r *SyncedRegistry) get(name string) *pb.Info {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()
	info := r.services[name]
	return clone.New(info).(*pb.Info)
}

func (r *SyncedRegistry) saveService(info *pb.Info) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()
	r.services[info.Namespace+":"+info.Name] = info
}

func (r *SyncedRegistry) deleteService(name string) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()
	delete(r.services, name)
}

func (r *SyncedRegistry) connected() {
	ctx := context.Background()
	stream, err := r.client.Listen(ctx, &pb.ListenRequest{})
	if err != nil {
		log.Printf("could not listen to registry server events: %s\n", err)
		return
	}
	defer stream.CloseSend()
	for !r.stop {
		event, err := stream.Recv()
		if err != nil {
			log.Printf("could not get event: %s\n", err)
			return
		}

		log.Printf("registry -> %s: %s\n", event.Type.String(), event.Name)

		switch event.Type {
		case pb.EventType_Updated, pb.EventType_Registered:
			r.saveService(event.Info)
		case pb.EventType_DeRegistered:
			r.deleteService(event.Name)
		}
	}
}

func (r *SyncedRegistry) disconnected() {
	r.services = nil
}

func NewSyncRegistry(server string, tlsConfig *tls.Config) *SyncedRegistry {
	return &SyncedRegistry{
		services:      map[string]*pb.Info{},
		tlsConfig:     tlsConfig,
		serverAddress: server,
		eventHandlers: map[string]pb.RegistryEventHandler{},
	}
}
