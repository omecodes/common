package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/zoenion/common/types"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"

	Cookies           = types.String("cookies")
	CtxCCookiesStore  = types.String("cookies")
	CtxResponseWriter = types.String("response_writer")
	Vars              = types.String("vars")
)

type (
	Server struct {
		*http.Server
	}

	Request struct {
		*http.Request
	}

	ResponseWriter interface {
		http.ResponseWriter
	}

	Resource struct {
		Mime      string
		BytesData []byte
		Stream    io.ReadCloser
	}

	RedirectURL struct {
		URL         string
		Code        int
		ContentType string
	}

	Content struct {
		Mime      string
		WithRange bool
		Offset    int64
		Length    int64
		Total     int64
		Data      interface{}
	}

	RequireAuth struct {
		Type  string
		Realm string
	}

	JSON map[string]interface{}

	Route struct {
		Name              string
		Path              string
		PathMatchesPrefix bool
		Method            string
		Handler           Handler
		Description       string
	}

	Routes []Route

	Header struct {
		Name  string
		Value string
	}

	Handler func(ctx context.Context, r *Request) (interface{}, error)

	Router interface {
		GetRoutes() Routes
	}

	Middleware func(handler Handler) Handler
)

func logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func final(ctx context.Context, h Handler, cookieStore *sessions.CookieStore, middlewareList ...Middleware) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := h
		for _, m := range middlewareList {
			handler = m(handler)
		}

		muxVars := mux.Vars(r)
		httpCtx := ctx

		//can add middleware stack here to enrich context before calling the handler
		httpCtx = context.WithValue(httpCtx, CtxResponseWriter, w)
		httpCtx = context.WithValue(httpCtx, CtxCCookiesStore, cookieStore)
		httpCtx = context.WithValue(httpCtx, Vars, muxVars)

		r = r.WithContext(httpCtx)
		status := http.StatusOK
		request := Request{r}
		result, err := handler(httpCtx, &request)
		if err != nil {
			status = errors.Parse(err.Error()).Code
			w.WriteHeader(status)
			return
		}

		if result != nil {
			writeResponse(w, status, result)
		}
	})
}

func writeResponse(w http.ResponseWriter, status int, data interface{}, headers ...Header) {
	for _, h := range headers {
		w.Header().Set(h.Name, h.Value)
	}

	if data == nil {
		return
	}

	if content, ok := data.(*Content); ok {
		writeContent(content, w)
		return
	}

	if reader, ok := data.(io.Reader); ok {
		writeReader(status, reader, w)
		return
	}

	if bytesArray, ok := data.([]byte); ok {
		w.WriteHeader(status)
		_, err := w.Write(bytesArray)
		if err != nil {
			log.Println("error wile writing http response content: ", err)
		}
		return
	}

	if resource, ok := data.(*Resource); ok {
		writeResource(status, resource, w)
		return
	}

	if redirectURL, ok := data.(*RedirectURL); ok {
		writeRedirect(redirectURL, w)
		return
	}

	if resuireAuth, ok := data.(*RequireAuth); ok {
		writeRequireAuth(resuireAuth, w)
		return
	}

	w.WriteHeader(status)
	bytes, _ := json.Marshal(data)
	_, err := w.Write(bytes)
	if err != nil {
		log.Println("error wile writing http response content: ", err)
	}
}

func writeContent(c *Content, w http.ResponseWriter) {
	w.Header().Set("Accept-Ranges", "bytes")
	if c.WithRange {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", c.Length))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", c.Offset, c.Offset+c.Length-1, c.Total))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", c.Total))
		w.WriteHeader(http.StatusOK)
	}

	switch d := c.Data.(type) {
	case io.Reader:
		buffer := make([]byte, 4096)
		for {
			n, err := d.Read(buffer)
			if n > 0 {
				_, err = w.Write(buffer[:n])
				if err != nil {
					log.Println("error wile writing http response content: ", err)
					break
				}
			}
			if err != nil {
				if io.EOF != err {
					log.Print(err)
				}
				break
			}
		}
	case []byte:
		_, err := w.Write(d)
		if err != nil {
			log.Println("error wile writing http response content: ", err)
		}
	}

	if closer, ok := c.Data.(io.Closer); ok {
		_ = closer.Close()
	}
}

func writeReader(status int, reader io.Reader, w http.ResponseWriter) {
	w.WriteHeader(status)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			_, err = w.Write(buffer[:n])
			if err != nil {
				log.Println("error wile writing http response content: ", err)
				return
			}
		}
		if err != nil && io.EOF != err {
			log.Print(err)
			return
		}
	}
}

func writeResource(status int, resource *Resource, w http.ResponseWriter) {
	w.Header().Set("Content-Type", resource.Mime)
	w.WriteHeader(status)

	if resource.Stream != nil {
		buffer := make([]byte, 1024)
		for {
			n, err := resource.Stream.Read(buffer)
			if n > 0 {
				_, _ = w.Write(buffer[:n])
			}

			if err != nil || n == 0 {
				if io.EOF != err {
					log.Print(err)
				}
				_ = resource.Stream.Close()
				return
			}
		}
	} else if resource.BytesData != nil {
		_, _ = w.Write(resource.BytesData)
	}
}

func writeRedirect(red *RedirectURL, w http.ResponseWriter) {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("<head>\n"))
	b.WriteString(fmt.Sprintf("\t<meta http-equiv=\"refresh\" content=\"0; URL=%s\" />\n", red.URL))
	b.WriteString(fmt.Sprintf("</head>"))
	contentBytes := []byte(b.String())

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Location", red.URL)
	w.WriteHeader(red.Code)
	_, _ = w.Write(contentBytes)
}

func writeRequireAuth(ar *RequireAuth, w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+ar.Realm+`"`)
	w.WriteHeader(http.StatusUnauthorized)
}

func StartServer(ctx context.Context, addr string, routes Routes, cookieStore *sessions.CookieStore, middlewareList ...Middleware) (*Server, error) {
	muxRouter := mux.NewRouter()
	for _, route := range routes {
		handler := final(ctx, route.Handler, cookieStore, middlewareList...)
		handler = logger(handler)
		if route.PathMatchesPrefix {
			muxRouter.PathPrefix(route.Path).HandlerFunc(handler.ServeHTTP).Methods(route.Method).Name(route.Name)
		} else {
			muxRouter.HandleFunc(route.Path, handler.ServeHTTP).Methods(route.Method).Name(route.Name)
		}
	}
	httpServer := http.Server{Handler: muxRouter, Addr: addr}
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			log.Println("failed to start HTTP server:", err)
		}
	}()
	return &Server{&httpServer}, nil
}

func StartSecureServer(ctx context.Context, addr string, routes Routes, certFile string, keyFile string, cookieStore *sessions.CookieStore, middlewareList ...Middleware) (*http.Server, error) {
	muxRouter := mux.NewRouter()
	for _, route := range routes {
		handler := final(ctx, route.Handler, cookieStore, middlewareList...)
		handler = logger(handler)
		if route.PathMatchesPrefix {
			muxRouter.PathPrefix(route.Path).HandlerFunc(handler.ServeHTTP).Methods(route.Method).Name(route.Name)
		} else {
			muxRouter.HandleFunc(route.Path, handler.ServeHTTP).Methods(route.Method).Name(route.Name)
		}
	}
	httpServer := &http.Server{Handler: muxRouter, Addr: addr}
	go func() {
		err := httpServer.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Println(err)
		}
	}()
	return httpServer, nil
}
