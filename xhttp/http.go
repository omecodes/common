package xhttp

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
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/types"
)

const (
	Cookies           = types.String("cookies")
	CtxCCookiesStore  = types.String("cookies")
	CtxResponseWriter = types.String("response_writer")
	HttpVars          = types.String("vars")
)

type (
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

	HttpHeader struct {
		Name  string
		Value string
	}

	RequestWrapper func(r *http.Request) *http.Request

	HttpMiddleware func(handler http.HandlerFunc) http.HandlerFunc
)

func logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}

func final(ctx context.Context, next http.HandlerFunc, cookieStore *sessions.CookieStore, middlewareList ...HttpMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := next
		for _, m := range middlewareList {
			handler = m(handler)
		}

		muxVars := mux.Vars(r)
		httpCtx := ctx
		//can add middleware stack here to enrich context before calling the handler
		httpCtx = context.WithValue(httpCtx, CtxResponseWriter, w)
		httpCtx = context.WithValue(httpCtx, CtxCCookiesStore, cookieStore)
		httpCtx = context.WithValue(httpCtx, HttpVars, muxVars)

		r = r.WithContext(httpCtx)
		handler(w, r)
	}
}

func ApiAccessMiddleware(key, secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || key != user && secret != password {
			WriteResponse(w, 401, &RequireAuth{
				Realm: "onfs-ds",
				Type:  "Basic",
			})
		}
		next(w, r)
	}
}

func HttpBasicMiddlewareStack(ctx context.Context, h http.HandlerFunc, cookieStore *sessions.CookieStore, middlewareList ...HttpMiddleware) http.HandlerFunc {
	handler := final(ctx, h.ServeHTTP, cookieStore)
	handler = logger(handler)
	return handler
}

func WriteError(w http.ResponseWriter, err error) {
	status := errors.Parse(err).Code
	w.WriteHeader(status)
}

func WriteResponse(w http.ResponseWriter, status int, data interface{}, headers ...HttpHeader) {
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
