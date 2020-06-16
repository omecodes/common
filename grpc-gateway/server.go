package grpc_gateway

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/omecodes/common/errors"
	ga "github.com/omecodes/common/grpc-authentication"
	gs "github.com/omecodes/common/grpc-session"
	"github.com/omecodes/common/netx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"strings"
)

type Mapper func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

type MuxWrapper func(mux *runtime.ServeMux) http.Handler

type gatewayConfigs struct {
	port        int
	mapper      Mapper
	muxWrappers []MuxWrapper
}

type Server struct {
	host         string
	options      options
	certificate  *x509.Certificate
	key          crypto.PrivateKey
	initialized  bool
	mux          *runtime.ServeMux
	grpcServer   *grpc.Server
	errorChannel chan error
	grpcAddress  string
	grpcListener net.Listener
	httpAddress  string
	httpListener net.Listener
	stopped      bool
}

func (s *Server) listenHttp() error {
	if s.options.gatewayConfigs != nil {

		address := fmt.Sprintf("%s:", s.host)
		if s.options.port > 0 {
			address = fmt.Sprintf("%s%d", address, s.options.port)
		}
		l, err := netx.Listen(address, s.options.listenOptions...)
		if err != nil {
			return err
		}
		s.httpListener = l

		if s.options.port > 0 {
			s.httpAddress = address
		} else {
			s.httpAddress = address + strings.Split(l.Addr().String(), ":")[1]
		}

		log.Printf("listening to HTTP traffic at %s", s.httpAddress)
	}
	return nil
}

func (s *Server) listenGRPC() error {

	address := fmt.Sprintf("%s:", s.host)
	if s.options.GRPC > 0 {
		address = fmt.Sprintf("%s%d", address, s.options.GRPC)
	}
	l, err := netx.Listen(address, s.options.listenOptions...)
	if err != nil {
		return err
	}
	s.httpListener = l

	if s.options.GRPC > 0 {
		s.grpcAddress = address
	} else {
		s.grpcAddress = address + strings.Split(l.Addr().String(), ":")[1]
	}

	log.Printf("listening to gRPC traffic at %s", s.grpcAddress)

	return nil
}

func (s *Server) init() error {
	if s.initialized {
		return nil
	}
	s.initialized = true

	err := s.listenGRPC()
	if err != nil {
		return err
	}

	err = s.listenHttp()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) startGRPC() {
	if !s.initialized {
		s.errorChannel <- errors.New("Init method must me called at least once")
		return
	}
	err := s.grpcServer.Serve(s.grpcListener)
	if err != nil {
		if !s.stopped {
			s.errorChannel <- err
		}
	}
}

func (s *Server) startHTTP() {
	if s.options.gatewayConfigs != nil {
		if !s.initialized {
			s.errorChannel <- errors.New("Init method must me called at least once")
			return
		}
		s.mux = runtime.NewServeMux(
			runtime.WithForwardResponseOption(gs.SetCookieFromGRPCMetadata),
		)

		var opts []grpc.DialOption

		var lopts netx.ListenOptions
		for _, o := range s.options.listenOptions {
			o(&lopts)
		}

		if lopts.Secure {
			if lopts.TLS != nil {
				cert := lopts.TLS.Certificates[0]
				c := credentials.NewServerTLSFromCert(&cert)
				opts = append(opts, grpc.WithTransportCredentials(c))
			} else {
				c, err := credentials.NewClientTLSFromFile(lopts.CertFilename, "")
				if err != nil {
					s.errorChannel <- err
					return
				}
				opts = append(opts, grpc.WithTransportCredentials(c))
			}
		}

		err := s.options.mapper(context.Background(), s.mux, s.grpcAddress, opts)
		if err != nil {
			s.errorChannel <- err
			return
		}

		var handler http.Handler
		if len(s.options.muxWrappers) > 0 {
			for _, wrapFunc := range s.options.muxWrappers {
				handler = wrapFunc(s.mux)
			}
		} else {
			handler = s.mux
		}

		err = http.Serve(s.httpListener, handler)
		if err != nil {
			if !s.stopped {
				s.errorChannel <- err
			}
		}
	}
}

func (s *Server) gRPCInterceptAuth(ctx context.Context) (context.Context, error) {
	basic, _ := grpc_auth.AuthFromMD(ctx, "basic")
	if basic != "" {
		bytes, err := base64.StdEncoding.DecodeString(basic)
		if err != nil {
			return nil, errors.Forbidden
		}

		parts := strings.Split(string(bytes), ":")
		if len(parts) != 2 {
			return nil, errors.Forbidden
		}

		user := parts[0]
		secret := parts[1]

		ctx = ga.ContextWithCredentials(ctx, &ga.Credentials{
			Username: user,
			Password: secret,
		})
		return ctx, nil
	}

	token, _ := grpc_auth.AuthFromMD(ctx, "bearer")
	if token != "" {
		ctx = ga.ContextWithOauth2Token(ctx, token)
	}

	return ctx, nil
}

func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return err
	}
	go s.startGRPC()
	go s.startHTTP()
	return nil
}

func (s *Server) Certificate() *x509.Certificate {
	return s.certificate
}

func (s *Server) Key() crypto.PrivateKey {
	return s.key
}

func (s *Server) GRPCServer() *grpc.Server {
	if s.grpcServer == nil {
		s.grpcServer = grpc.NewServer(
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_opentracing.StreamServerInterceptor(),
				grpc_prometheus.StreamServerInterceptor,
				grpc_auth.StreamServerInterceptor(s.gRPCInterceptAuth),
				grpc_recovery.StreamServerInterceptor(),
			)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(),
				grpc_prometheus.UnaryServerInterceptor,
				grpc_auth.UnaryServerInterceptor(s.gRPCInterceptAuth),
				grpc_recovery.UnaryServerInterceptor(),
			)),
		)
	}
	return s.grpcServer
}

func (s *Server) Stop() {
	s.stopped = true

	if s.httpListener != nil {
		_ = s.httpListener.Close()
	}

	if s.grpcListener != nil {
		_ = s.httpListener.Close()
	}
}

func (s *Server) GatewayAddress() string {
	return s.httpAddress
}

func (s *Server) GRPCAddress() string {
	return s.grpcAddress
}

func (s *Server) Errors() chan error {
	return s.errorChannel
}

func New(host string, opts ...Option) *Server {
	s := &Server{
		errorChannel: make(chan error, 2),
	}

	for _, o := range opts {
		o(&s.options)
	}

	return s
}
