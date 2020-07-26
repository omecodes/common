package pb

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	crypto2 "github.com/omecodes/common/crypto"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"sync"
)

type poolOptions struct {
	caCertFilename string
	certFilename   string
	keyFilename    string
	keyPassword    []byte
}

type PoolOption func(p *poolOptions)

func CACert(filename string) PoolOption {
	return func(p *poolOptions) {
		p.caCertFilename = filename
	}
}

func Cert(filename string) PoolOption {
	return func(p *poolOptions) {
		p.certFilename = filename
	}
}

func Key(filename string, password []byte) PoolOption {
	return func(p *poolOptions) {
		p.keyPassword = password
		p.keyFilename = filename
	}
}

func NewConnectionPool(registry Registry, opts ...PoolOption) ConnectionPool {
	p := &pool{
		registry: registry,
		dialer:   map[string]Dialer{},
	}

	p.options = new(poolOptions)
	for _, opt := range opts {
		opt(p.options)
	}

	return p
}

type ConnectionPool interface {
	Dialer(name string) (Dialer, error)
	Connection(name string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type pool struct {
	sync.Mutex
	options *poolOptions
	caCert  *x509.Certificate
	cert    *x509.Certificate
	key     crypto.PrivateKey

	dialer   map[string]Dialer
	registry Registry
}

func (p *pool) Dialer(name string) (Dialer, error) {
	p.Lock()
	defer p.Unlock()

	dialer, found := p.dialer[name]
	if !found {
		info, err := p.registry.GetService(name)
		if err != nil {
			return nil, err
		}

		for _, node := range info.Nodes {

			if node.Protocol == Protocol_Grpc {
				var opts []grpc.DialOption
				tc, err := p.getTLSConfig(node)
				if err != nil {
					log.Error("failed to get TLS config", err)
					return nil, err
				}

				if tc != nil {
					opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tc)))
				} else {
					opts = append(opts, grpc.WithInsecure())
				}

				dialer := NewDialer(node.Address, opts...)
				p.dialer[name] = dialer
				return dialer, nil
			}
		}
		return nil, errors.NotFound

	} else {
		return dialer, nil
	}
}

func (p *pool) Connection(name string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialer, err := p.Dialer(name)
	if err != nil {
		return nil, err
	}
	return dialer.Dial(opts...)
}

func (p *pool) getTLSConfig(node *Node) (*tls.Config, error) {
	if node.Security == Security_MutualTLS {
		return p.mutualTLS(node)
	}

	if node.Security == Security_TLS {
		return p.tlsConfig(node)
	}

	return nil, nil
}

func (p *pool) mutualTLS(node *Node) (*tls.Config, error) {
	caCert, err := p.getCaCert()
	if err != nil {
		return nil, err
	}
	CAPool := x509.NewCertPool()
	CAPool.AddCert(caCert)

	cert, err := p.getCert()
	if err != nil {
		return nil, err
	}

	key, err := p.getPrivateKey()
	if err != nil {
		return nil, err
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
	}

	return &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}

func (p *pool) tlsConfig(node *Node) (*tls.Config, error) {
	caCert, err := p.getCaCert()
	if err != nil {
		return nil, err
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(caCert)

	return &tls.Config{
		RootCAs: CAPool,
	}, nil
}

func (p *pool) getCaCert() (*x509.Certificate, error) {
	var err error
	if p.caCert == nil && p.options.caCertFilename != "" {
		p.caCert, err = crypto2.LoadCertificate(p.options.caCertFilename)
	}
	return p.caCert, err
}

func (p *pool) getCert() (*x509.Certificate, error) {
	var err error
	if p.cert == nil && p.options.certFilename != "" {
		p.cert, err = crypto2.LoadCertificate(p.options.certFilename)
	}
	return p.cert, err
}

func (p *pool) getPrivateKey() (crypto.PrivateKey, error) {
	var err error
	if p.key == nil && p.options.keyFilename != "" {
		p.key, err = crypto2.LoadPrivateKey(p.options.keyPassword, p.options.keyFilename)
	}
	return p.key, err
}

type Dialer interface {
	Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type dialer struct {
	address        string
	wrapped        *grpc.ClientConn
	defaultOptions []grpc.DialOption
}

func (g *dialer) Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if g.wrapped == nil || g.wrapped.GetState() != connectivity.Ready {
		if g.wrapped != nil {
			_ = g.wrapped.Close()
		}
		var err error
		g.wrapped, err = grpc.Dial(g.address, g.defaultOptions...)
		if err != nil {
			return nil, err
		}
	}
	return g.wrapped, nil
}

func NewDialer(addr string, opts ...grpc.DialOption) *dialer {
	return &dialer{
		address:        addr,
		defaultOptions: opts,
	}
}
