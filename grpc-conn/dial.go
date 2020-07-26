package grpc_conn

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"sync"
)

var globalMutex = &sync.Mutex{}
var pool = map[string]Dialer{}

func init() {

}

type Dialer interface {
	Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type dialer struct {
	address string
	wrapped *grpc.ClientConn
	options []grpc.DialOption
}

func (g *dialer) Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if g.wrapped == nil || g.wrapped.GetState() != connectivity.Ready {
		if g.wrapped != nil {
			_ = g.wrapped.Close()
		}
		var err error
		g.wrapped, err = grpc.Dial(g.address, g.options...)
		if err != nil {
			return nil, err
		}
	}
	return g.wrapped, nil
}

func NewDialer(addr string, opts ...grpc.DialOption) *dialer {
	return &dialer{
		address: addr,
		options: opts,
	}
}

func Get(address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	dialer, found := pool[address]
	if found {
		return dialer.Dial()
	}

	dialer = NewDialer(address, opts...)
	pool[address] = dialer

	return dialer.Dial(opts...)
}
