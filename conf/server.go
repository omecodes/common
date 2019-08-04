package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/futils"
	http_helper "github.com/zoenion/common/http-helper"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Server struct {
	filename string
	address  string
	sync.Mutex
	authStore       map[string]string
	requireResponse *http_helper.RequireAuth
	data            Map
	apiSrv          *http.Server
}

func (s *Server) Start() (err error) {

	router := mux.NewRouter()

	router.PathPrefix(RouteSet).HandlerFunc(s.apiSet).Methods(http.MethodPost).Name("configs.set")
	router.PathPrefix(RouteGet).HandlerFunc(s.apiGet).Methods(http.MethodGet).Name("configs.get")
	router.PathPrefix(RouteDel).HandlerFunc(s.apiDel).Methods(http.MethodDelete).Name("configs.del")

	s.apiSrv = &http.Server{
		Addr:    s.address,
		Handler: router,
	}

	err = s.apiSrv.ListenAndServe()
	if err == nil {
		log.Println(fmt.Sprintf("API SERVER: http://%s", s.apiSrv.Addr))
		log.Println("API Accesses:")
		for k, v := range s.authStore {
			log.Printf("%s:%s\n", k, v)
		}
	}
	return err
}

func (s *Server) Stop() error {
	return s.apiSrv.Shutdown(context.Background())
}

func (s *Server) apiGet(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	path := strings.Replace(r.URL.Path, RouteGet, "", 1)
	s.data.GetConf(path)
}

func (s *Server) apiSet(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	path := strings.Replace(r.URL.Path, RouteSet, "", 1)

	var m Map
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http_helper.WriteError(w, err)
		return
	}

	s.data.Set(path, m)
	if err := s.data.Save(s.filename, os.ModePerm); err != nil {
		http_helper.WriteError(w, err)
	}
}

func (s *Server) apiDel(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	path := strings.Replace(r.URL.Path, RouteDel, "", 1)
	s.data.Del(path)
	if err := s.data.Save(s.filename, os.ModePerm); err != nil {
		http_helper.WriteError(w, err)
	}
}

func NewServer(cfgFilename string) (*Server, error) {
	s := new(Server)
	var err error
	cfgFilename, err = filepath.Abs(cfgFilename)
	if err != nil {
		return nil, err
	}

	conf := Map{}
	err = Load(cfgFilename, &conf)
	if err != nil {
		log.Println(err)
		log.Printf("dir must contain conf.json file formatted as below \n%s")
		return nil, errors.New("could not load configs file at " + cfgFilename)
	}

	address, ok := conf.GetString("address")
	if !ok {
		return nil, errors.New("invalid configs file")
	}
	s.address = address

	storeFilename, ok := conf.GetString("store")
	if !ok {
		return nil, errors.New("invalid configs file")
	}

	s.filename = storeFilename
	s.requireResponse = &http_helper.RequireAuth{
		Type:  "Basic",
		Realm: "conf-server",
	}

	if authStore := conf.GetConf("auth"); authStore != nil {
		s.authStore = map[string]string{}
		for k, v := range authStore {
			s.authStore[k] = v.(string)
		}
	}
	s.data = Map{}
	if futils.FileExists(s.filename) {
		err = Load(s.filename, &s.data)
	}
	return s, err
}
