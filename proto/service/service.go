package servicepb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"log"
	"sync"
	"time"
)

type RegistryEventHandler interface {
	Handle(*Event)
}

func NewClient(serverAddr string, opts ...grpc.DialOption) (RegistryClient, error) {
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, err
	}
	return NewRegistryClient(conn), nil
}

type SyncRegistry struct {
	servicesLock sync.Mutex
	handlersLock sync.Mutex
	services     map[string]*Info
	client       RegistryClient

	serverCert    *x509.Certificate
	serverAddress string
	stop          bool
	conn          *grpc.ClientConn
	eventHandlers []RegistryEventHandler
}

func (r *SyncRegistry) Connect() error {
	if r.conn != nil && r.conn.GetState() == connectivity.Ready {
		return nil
	}

	var opts []grpc.DialOption

	if r.serverCert != nil {
		tlsCert := &tls.Certificate{Certificate: [][]byte{r.serverCert.Raw}}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(tlsCert)))
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
		}
	}

	r.client = NewRegistryClient(r.conn)
	go r.connected()
	return nil
}

func (r *SyncRegistry) Disconnect() error {
	r.stop = true
	r.disconnected()
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *SyncRegistry) Register(i *Info) (string, error) {
	err := r.Connect()
	if err != nil {
		return "", fmt.Errorf("could not connect to server: %s", err)
	}
	rsp, err := r.client.Register(context.Background(), &RegisterRequest{})
	if err != nil {
		return "", err
	}
	return rsp.RegistryId, nil
}

func (r *SyncRegistry) Deregister(id string) error {
	err := r.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to server: %s", err)
	}

	_, err = r.client.Deregister(context.Background(), &DeregisterRequest{RegistryId: id})
	return err
}

func (r *SyncRegistry) GetAddress(namespace string, name string, protocol string) (string, error) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()

	for _, s := range r.services {
		if s.Namespace == namespace && s.Name == name {
			for _, n := range s.Nodes {
				if n.Protocol == protocol {
				}
			}
		}
	}
	return "", nil
}

func (r *SyncRegistry) AddEventHandler(h RegistryEventHandler) {
	r.handlersLock.Lock()
	defer r.handlersLock.Lock()
	r.eventHandlers = append(r.eventHandlers, h)
}

func (r *SyncRegistry) dispatchEvent(e Event) {
	r.handlersLock.Lock()
	r.handlersLock.Unlock()

	for _, handler := range r.eventHandlers {
		handler.Handle(&e)
	}
}

func (r *SyncRegistry) saveService(info *Info) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()
	r.services[info.Namespace+"::"+info.Name] = info
}

func (r *SyncRegistry) deleteService(name string) {
	delete(r.services, name)
}

func (r *SyncRegistry) connected() {
	ctx := context.Background()
	stream, err := r.client.Listen(ctx, &ListenRequest{})
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
		case EventType_Updated, EventType_Registered:
			r.saveService(event.Info)
		case EventType_DeRegistered:
			r.deleteService(event.Name)
		}
	}
}

func (r *SyncRegistry) disconnected() {
	r.services = nil
}

func NewSyncRegistry(server string, cert *x509.Certificate) *SyncRegistry {
	return &SyncRegistry{
		serverCert:    cert,
		serverAddress: server,
	}
}
