package conf

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pydio/cells/common/config"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/futils"
	"github.com/zoenion/common/gateway"
	http_helper "github.com/zoenion/common/http-helper"
	"github.com/zoenion/common/proto/config"
	"google.golang.org/grpc"
	"os"
	"path/filepath"
	"sync"
)

type ServerInfo struct {
	GRPC                    string
	HTTP                    string
	TLS                     *tls.Config
	accessKey, accessSecret string
	Filename                string
}

type Server struct {
	cfgLock  sync.Mutex
	authLock sync.Mutex
	info     *ServerInfo
	address  string
	data     Map
	gateway  *gateway.Gateway
}

func (s *Server) Set(ctx context.Context, in *configpb.SetRequest) (*configpb.SetResponse, error) {
	s.cfgLock.Lock()
	defer s.cfgLock.Unlock()

	cfg := config.Map{}
	err := json.Unmarshal(in.Value, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %s", err)
	}

	s.data.Set(in.Key, cfg)
	rsp := &configpb.SetResponse{}

	err = s.data.Save(s.info.Filename, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("could not save configs: %s", err)
	}
	return rsp, err
}

func (s *Server) Get(ctx context.Context, in *configpb.GetRequest) (*configpb.GetResponse, error) {
	s.cfgLock.Lock()
	defer s.cfgLock.Unlock()

	cfg := s.data.GetConf(in.Key)
	if cfg != nil {
		return nil, errors.NotFound
	}
	var err error
	rsp := &configpb.GetResponse{}
	rsp.Value, err = json.Marshal(cfg)

	return rsp, err
}

func (s *Server) Delete(ctx context.Context, in *configpb.DeleteRequest) (*configpb.DeleteResponse, error) {
	s.cfgLock.Lock()
	defer s.cfgLock.Unlock()

	s.data.Del(in.Key)
	err := s.data.Save(s.info.Filename, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("could not save configs: %s", err)
	}
	return &configpb.DeleteResponse{}, err
}

func (s *Server) Start() (err error) {
	return s.gateway.Start()
}

func (s *Server) Stop() error {
	s.gateway.Stop()
	return nil
}

func NewServer(name string, info *ServerInfo) (*Server, error) {
	s := new(Server)
	var err error

	info.Filename, err = filepath.Abs(info.Filename)
	if err != nil {
		return nil, err
	}

	s.data = Map{}
	if futils.FileExists(info.Filename) {
		err = Load(info.Filename, &s.data)
	}

	gc := &gateway.Config{
		Name: name,
		Tls:  info.TLS,
		HTTP: &gateway.HTTP{
			Address:        info.HTTP,
			WireGRPCFunc:   configpb.RegisterConfigHandlerFromEndpoint,
			MiddlewareList: []http_helper.HttpMiddleware{},
		},
		GRPC: &gateway.GRPC{
			Address: info.GRPC,
			RegisterHandlerFunc: func(srv *grpc.Server) {
				configpb.RegisterConfigServer(srv, s)
			},
		},
	}
	s.gateway = gateway.New(gc)
	return s, err
}
