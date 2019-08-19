package servicepb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

type Registry struct {
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

func (r *Registry) Connect() error {
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

	for !r.stop || r.conn != nil && r.conn.GetState() != connectivity.Ready {
		var err error
		r.conn, err = grpc.Dial(r.serverAddress, opts...)
		if err != nil {
			log.Printf("connection to registry server failed: %s", err)
			<-time.After(time.Second * 3)
		}
	}

	r.client = NewRegistryClient(r.conn)
	go r.connected()
	return nil
}

func (r *Registry) Disconnect() error {
	r.stop = true
	r.disconnected()
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *Registry) AddEventHandler(h RegistryEventHandler) {
	r.servicesLock.Lock()
	defer r.servicesLock.Lock()
}

func (r *Registry) GetAddress(namespace string, name string, protocol string) (string, error) {
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

func (r *Registry) saveService(info *Info) {
	r.servicesLock.Lock()
	defer r.servicesLock.Unlock()
	r.services[info.Namespace+"::"+info.Name] = info
}

func (r *Registry) deleteService(name string) {
	delete(r.services, name)
}

func (r *Registry) connected() {
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

func (r *Registry) disconnected() {
	r.services = nil
}

func NewRegistry(server string, cert *x509.Certificate) *Registry {
	return &Registry{
		serverCert:    cert,
		serverAddress: server,
	}
}
