package authpb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/zoenion/common/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"log"
	"sync"
	"time"
)

type SyncedJwtStoreClient struct {
	localStoreMutex sync.Mutex
	localStore      map[string]*JwtInfo

	serverAddress string
	serverCert    *x509.Certificate
	conn          *grpc.ClientConn
}

func (s *SyncedJwtStoreClient) connect() (err error) {
	if s.conn != nil && s.conn.GetState() == connectivity.Ready {
		return nil
	}
	if s.serverCert != nil {
		tlsCert := &tls.Certificate{Certificate: [][]byte{s.serverCert.Raw}}
		s.conn, err = grpc.Dial(s.serverAddress, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(tlsCert)))
	} else {
		s.conn, err = grpc.Dial(s.serverAddress, grpc.WithInsecure())
	}

	if err == nil {
		go s.connected()
	}
	return err
}

func (s *SyncedJwtStoreClient) connected() {
	client := NewJWTStoreClient(s.conn)
	stream, err := client.Listen(context.Background(), &ListenRequest{})
	if err != nil {
		log.Printf("could not listen to jwt events: %s\n", err)
		return
	}

	defer stream.CloseSend()
	for {
		event, err := stream.Recv()
		if err != nil {
			log.Printf("received error while read jwt events stream: %s\n", err)
			return
		}

		switch event.Action {
		case EventAction_Save:
			s.saveJwtInfo(event.Info)
		case EventAction_Delete:
			s.deleteJwtInfo(event.Info.Jti)
		}
	}
}

func (s *SyncedJwtStoreClient) saveJwtInfo(i *JwtInfo) {
	s.localStoreMutex.Lock()
	defer s.localStoreMutex.Unlock()
	s.localStore[i.Jti] = i
}

func (s *SyncedJwtStoreClient) deleteJwtInfo(jti string) {
	s.localStoreMutex.Lock()
	defer s.localStoreMutex.Unlock()
	delete(s.localStore, jti)
}

func (s *SyncedJwtStoreClient) getJwtState(jti string) (JWTState, error) {
	s.localStoreMutex.Lock()
	s.localStoreMutex.Unlock()

	info, found := s.localStore[jti]
	if !found {
		return 0, errors.NotFound
	}

	now := time.Now().Unix()
	if info.Nbf > now {
		return JWTState_NOT_EFFECTIVE, errors.New("jwt not effective")
	}

	if info.Exp < now {
		return JWTState_EXPIRED, errors.New("jwt expired")
	}

	return JWTState_VALID, nil
}

func (s *SyncedJwtStoreClient) State(jti string) (JWTState, error) {
	return s.getJwtState(jti)
}

func NewSyncJwtStoreClient(server string, cert *x509.Certificate) {

}

type JwtFeeder struct {
	serverAddress string
	serverCert    *x509.Certificate

	conn   *grpc.ClientConn
	stream JWTStore_FeedClient
}

func (f *JwtFeeder) connect() (err error) {
	if f.conn != nil && f.conn.GetState() == connectivity.Ready {
		return nil
	}
	if f.serverCert != nil {
		tlsCert := &tls.Certificate{Certificate: [][]byte{f.serverCert.Raw}}
		f.conn, err = grpc.Dial(f.serverAddress, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(tlsCert)))
	} else {
		f.conn, err = grpc.Dial(f.serverAddress, grpc.WithInsecure())
	}
	return err
}

func (f *JwtFeeder) AddJwt(info *JwtInfo) (err error) {
	if err = f.connect(); err != nil {
		return
	}

	if f.stream == nil {
		client := NewJWTStoreClient(f.conn)
		f.stream, err = client.Feed(context.Background())
		if err != nil {
			return
		}
	}

	return f.stream.Send(&JwtEvent{
		Info:   info,
		Action: EventAction_Save,
	})
}

func (f *JwtFeeder) Revoke(jti string) (err error) {
	if err = f.connect(); err != nil {
		return
	}

	if f.stream == nil {
		client := NewJWTStoreClient(f.conn)
		f.stream, err = client.Feed(context.Background())
		if err != nil {
			return
		}
	}

	return f.stream.Send(&JwtEvent{
		Action: EventAction_Delete,
		Info:   &JwtInfo{Jti: jti},
	})
}

func NewJwtFeeder(serverAddress string, serverCert *x509.Certificate) *JwtFeeder {
	return &JwtFeeder{
		serverAddress: serverAddress,
		serverCert:    serverCert,
	}
}
