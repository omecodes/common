package jwt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/zoenion/common/data"
	"github.com/zoenion/common/errors"
	authpb "github.com/zoenion/common/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"log"
	"sync"
	"time"
)

type SyncStore struct {
	sync.Mutex
	store                     data.Dict
	serverAddress             string
	serverCert                *x509.Certificate
	conn                      *grpc.ClientConn
	jwtRevokedHandlerFuncList map[string]RevokedHandlerFunc
}

func (s *SyncStore) connect() (err error) {
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

func (s *SyncStore) connected() {
	client := authpb.NewJWTStoreClient(s.conn)
	stream, err := client.Listen(context.Background(), &authpb.ListenRequest{})
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
		case authpb.EventAction_Save:
			_ = s.saveJwtInfo(event.Info)
		case authpb.EventAction_Delete:
			_ = s.deleteJwtInfo(event.Info.Jti)
		}
	}
}

func (s *SyncStore) saveJwtInfo(i *authpb.JwtInfo) error {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return s.store.Set(i.Jti, marshaled)
}

func (s *SyncStore) deleteJwtInfo(jti string) error {
	return s.store.Del(jti)
}

func (s *SyncStore) getJwtState(jti string) (authpb.JWTState, error) {
	infoBytes, err := s.store.Get(jti)
	if err != nil {
		return 0, err
	}

	info := new(authpb.JwtInfo)
	err = json.Unmarshal(infoBytes, info)
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	if info.Nbf > now {
		return authpb.JWTState_NOT_EFFECTIVE, errors.New("jwt not effective")
	}

	if info.Exp < now {
		return authpb.JWTState_EXPIRED, errors.New("jwt expired")
	}

	return authpb.JWTState_VALID, nil
}

func (s *SyncStore) State(jti string) (authpb.JWTState, error) {
	return s.getJwtState(jti)
}

func (s *SyncStore) AddJwtRevokedEventHandler(f RevokedHandlerFunc) string {
	s.Lock()
	defer s.Unlock()
	id := uuid.New().String()
	s.jwtRevokedHandlerFuncList[id] = f
	return id
}

func (s *SyncStore) DeleteJwtRevokedEventHandler(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.jwtRevokedHandlerFuncList, id)
}

func NewSyncJwtStore(server string, cert *x509.Certificate, store data.Dict) *SyncStore {
	return &SyncStore{
		serverAddress:             server,
		serverCert:                cert,
		store:                     store,
		jwtRevokedHandlerFuncList: map[string]RevokedHandlerFunc{},
	}
}

type Feeder struct {
	serverAddress string
	serverCert    *x509.Certificate
	conn          *grpc.ClientConn
	stream        authpb.JWTStore_FeedClient
}

func (f *Feeder) connect() (err error) {
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

func (f *Feeder) Register(info *authpb.JwtInfo) (err error) {
	if err = f.connect(); err != nil {
		return
	}

	if f.stream == nil {
		client := authpb.NewJWTStoreClient(f.conn)
		f.stream, err = client.Feed(context.Background())
		if err != nil {
			return
		}
	}

	return f.stream.Send(&authpb.JwtEvent{
		Info:   info,
		Action: authpb.EventAction_Save,
	})
}

func (f *Feeder) Revoke(jti string) (err error) {
	if err = f.connect(); err != nil {
		return
	}

	if f.stream == nil {
		client := authpb.NewJWTStoreClient(f.conn)
		f.stream, err = client.Feed(context.Background())
		if err != nil {
			return
		}
	}

	return f.stream.Send(&authpb.JwtEvent{
		Action: authpb.EventAction_Delete,
		Info:   &authpb.JwtInfo{Jti: jti},
	})
}

func NewJwtFeeder(serverAddress string, serverCert *x509.Certificate) *Feeder {
	return &Feeder{
		serverAddress: serverAddress,
		serverCert:    serverCert,
	}
}
