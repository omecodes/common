package front

import (
	"github.com/gorilla/mux"
	"github.com/omecodes/common/futils"
	"github.com/omecodes/common/web/app"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	staticRoutePrefix = "/static"
	webAppRoutePrefix = "/app/{name}"
	apiRoutePrefix    = "/service/{service}/api"
)

type Server struct {
	sync.Mutex
	addr        string
	opts        Options
	initialized bool
	handler     http.Handler
	router      *mux.Router
	webapps     *app.Webapp
	errorChan   chan error
}

func New(addr string, opts ...Option) *Server {
	srv := new(Server)
	for _, opt := range opts {
		opt(&srv.opts)
	}
	if addr == "" {
		addr = ":"
	}
	srv.addr = addr
	srv.errorChan = make(chan error)
	return srv
}

func (srv *Server) init() error {
	srv.Lock()
	defer srv.Unlock()

	if srv.initialized {
		return nil
	}
	srv.initialized = true

	if srv.opts.staticFilesRoutePrefix == "" {
		srv.opts.staticFilesRoutePrefix = staticRoutePrefix
	}

	if srv.opts.webAppsRoutePrefix == "" {
		srv.opts.webAppsRoutePrefix = webAppRoutePrefix
	}

	srv.router = mux.NewRouter().StrictSlash(true)

	if srv.opts.staticFolder != "" {
		srv.opts.staticFolder, _ = filepath.Abs(srv.opts.staticFolder)
		if !futils.FileExists(srv.opts.staticFolder) {
			err := os.MkdirAll(srv.opts.staticFolder, os.ModePerm)
			if err != nil {
				return err
			}
		}
		router := srv.router.PathPrefix(staticRoutePrefix).Subrouter()
		router.Name("statics").Methods(http.MethodGet, http.MethodPost).Handler(
			http.StripPrefix(staticRoutePrefix, http.FileServer(http.Dir(srv.opts.staticFolder))))
	}

	if srv.opts.webAppsFolder != "" {
		srv.opts.webAppsFolder, _ = filepath.Abs(srv.opts.webAppsFolder)
		if !futils.FileExists(srv.opts.webAppsFolder) {
			err := os.MkdirAll(srv.opts.webAppsFolder, os.ModePerm)
			if err != nil {
				return err
			}
		}
		srv.webapps = app.NewFolder(srv.opts.webAppsFolder, srv.opts.i18n)
		router := srv.router.PathPrefix(webAppRoutePrefix).Subrouter()
		router.Name("webapps").Methods(http.MethodGet).Handler(srv.webapps)
	}

	if srv.opts.apiHandler != nil {
		router := srv.router.PathPrefix(apiRoutePrefix + "/").Subrouter()
		router.Name("api").
			Methods(http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut).
			Handler(srv.opts.apiHandler)
	}

	return nil
}

func (srv *Server) Start() {
	go func() {
		err := srv.init()
		if err != nil {
			srv.errorChan <- err
		}
		if srv.opts.tls {
			srv.errorChan <- http.ListenAndServeTLS(srv.addr, srv.opts.certFilename, srv.opts.keyFilename, srv.handler)
		} else {
			srv.errorChan <- http.ListenAndServe(srv.addr, srv.router)
		}
	}()
}

func (srv *Server) Error() error {
	err, _ := <-srv.errorChan
	return err
}
