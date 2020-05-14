package grpc_gateway

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
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
	"github.com/iancoleman/strcase"
	crypto2 "github.com/zoenion/common/crypto"
	"github.com/zoenion/common/errors"
	ga "github.com/zoenion/common/grpc-authentication"
	gs "github.com/zoenion/common/grpc-session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Mapper func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

type MuxWrapper func(mux *runtime.ServeMux) http.Handler

type Configs struct {
	Name                string
	CertificateFilename string
	KeyFilename         string
	SelfSigned          bool
	Ip                  string
	Domain              string
	GRPCPort            int
	HTTPPort            int
	Mapper              Mapper
	MuxWrapper          MuxWrapper
}

type Server struct {
	configs      *Configs
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
	tc := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{s.certificate.Raw},
				PrivateKey:  s.key,
			},
		},
	}

	if s.configs.SelfSigned {
		pool := x509.NewCertPool()
		pool.AddCert(s.certificate)
		tc.ClientCAs = pool
	}

	address := fmt.Sprintf("%s:", s.configs.Domain)
	if s.configs.HTTPPort > 0 {
		address = fmt.Sprintf("%s%d", address, s.configs.HTTPPort)
	}
	l, err := tls.Listen("tcp", address, tc)
	if err != nil {
		return err
	}

	s.httpListener = l
	s.httpAddress = l.Addr().String()

	log.Printf("[%s-server] listening to HTTP traffic at %s", s.configs.Name, s.httpAddress)
	return nil
}

func (s *Server) listenGRPC() error {
	tc := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{s.certificate.Raw},
				PrivateKey:  s.key,
			},
		},
	}

	if s.configs.SelfSigned {
		pool := x509.NewCertPool()
		pool.AddCert(s.certificate)
		tc.ClientCAs = pool
	}

	address := fmt.Sprintf("%s:", s.configs.Domain)
	if s.configs.GRPCPort > 0 {
		address = fmt.Sprintf("%s%d", address, s.configs.GRPCPort)
	}

	l, err := tls.Listen("tcp", address, tc)
	if err != nil {
		return err
	}

	s.grpcListener = l
	s.grpcAddress = l.Addr().String()
	log.Printf("[%s-server] listening to gRPC traffic at %s\n", s.configs.Name, s.grpcAddress)

	return nil
}

func (s *Server) init() error {
	if s.initialized {
		return nil
	}
	s.initialized = true

	err := s.initCertAndKey()
	if err != nil {
		return err
	}

	err = s.listenGRPC()
	if err != nil {
		return err
	}

	err = s.listenHttp()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) initCertAndKey() error {
	var err error
	s.certificate, err = crypto2.LoadCertificate(s.configs.CertificateFilename)
	if err == nil {
		s.key, err = crypto2.LoadPrivateKey(nil, s.configs.KeyFilename)
	}

	if err == nil {
		inIP := net.ParseIP(s.configs.Ip)
		foundIP := false
		for _, ip := range s.certificate.IPAddresses {
			if ip.Equal(inIP) {
				foundIP = true
			}
		}

		if !foundIP || time.Now().After(s.certificate.NotAfter) || time.Now().Before(s.certificate.NotBefore) {
			err = errors.New("certificate is not valid")
		}
	}

	if err != nil {
		if s.configs.SelfSigned {
			s.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				return errors.Errorf("could not generate key pair: %s", err)
			}
			pub := s.key.(*ecdsa.PrivateKey).PublicKey

			caCertTemplate := &crypto2.CertificateTemplate{
				Organization:      "Ome",
				Name:              strcase.ToLowerCamel(s.configs.Name),
				Domains:           []string{s.configs.Domain},
				IPs:               []net.IP{net.ParseIP(s.configs.Ip)},
				Expiry:            time.Hour * 24 * 370,
				PublicKey:         &pub,
				SignerPrivateKey:  s.key,
				SignerCertificate: s.certificate,
			}
			s.certificate, err = crypto2.GenerateCACertificate(caCertTemplate)
			if err != nil {
				return errors.Errorf("could not generate application certificate: %s", err)
			}

			_ = crypto2.StorePrivateKey(s.key, nil, s.configs.KeyFilename)
			_ = crypto2.StoreCertificate(s.certificate, s.configs.CertificateFilename, os.ModePerm)
		} else {
			return errors.Errorf("could not load certificate/key pair for tls:", err)
		}
	}

	return err
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
	if !s.initialized {
		s.errorChannel <- errors.New("Init method must me called at least once")
		return
	}

	s.mux = runtime.NewServeMux(
		runtime.WithForwardResponseOption(gs.SetCookieFromGRPCMetadata),
	)

	var opts []grpc.DialOption
	c, err := credentials.NewClientTLSFromFile(s.configs.CertificateFilename, "")
	if err != nil {
		s.errorChannel <- err
		return
	}
	opts = append(opts, grpc.WithTransportCredentials(c))

	err = s.configs.Mapper(context.Background(), s.mux, s.grpcAddress, opts)
	if err != nil {
		s.errorChannel <- err
		return
	}

	var handler http.Handler
	if s.configs.MuxWrapper != nil {
		handler = s.configs.MuxWrapper(s.mux)
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

	bearer, _ := grpc_auth.AuthFromMD(ctx, "bearer")
	if bearer != "" {
		ctx = ga.ContextWithJWT(ctx, bearer)
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

func NewServer(cfg *Configs) *Server {
	return &Server{
		configs:      cfg,
		errorChannel: make(chan error, 2),
	}
}
