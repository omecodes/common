package xhttp

import (
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/types"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	CtxResponseWriter = types.String("response_writer")
)

type (
	Resource struct {
		Mime      string
		BytesData []byte
		Stream    io.ReadCloser
	}

	RedirectURL struct {
		URL         string
		Params      map[string]string
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

func WriteError(w http.ResponseWriter, err error) {
	status := errors.HttpStatus(err)
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
		WriteData(w, status, reader)
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
		writeRedirect(w, redirectURL)
		return
	}

	if requireAuth, ok := data.(*RequireAuth); ok {
		writeRequireAuth(requireAuth, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	bytes, _ := json.Marshal(data)
	_, err := w.Write(bytes)
	if err != nil {
		log.Println("error wile writing http response content: ", err)
	}
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	WriteResponse(w, status, data, HttpHeader{Name: "Content-Type", Value: "application/json"})
}

func Redirect(w http.ResponseWriter, url *RedirectURL) {
	writeRedirect(w, url)
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
		done := false
		buffer := make([]byte, 1024)
		for !done {
			n, err := d.Read(buffer)
			done = err == io.EOF
			if !done && err != nil {
				return
			}
			_, err = w.Write(buffer[:n])
			if err != nil {
				log.Println("[xhttp]:\tcould not write response:", err)
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

func WriteData(w http.ResponseWriter, status int, reader io.Reader) {
	w.WriteHeader(status)
	done := false
	buffer := make([]byte, 1024)
	for !done {
		n, err := reader.Read(buffer)
		done = err == io.EOF
		if !done && err != nil {
			return
		}
		_, err = w.Write(buffer[:n])
		if err != nil {
			log.Println("[xhttp]:\tcould not write response:", err)
		}
	}
}

func WriteContent(w http.ResponseWriter, contentType string, size int64, reader io.Reader) {

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.WriteHeader(http.StatusOK)
	done := false
	buffer := make([]byte, 1024)
	for !done {
		n, err := reader.Read(buffer)
		done = err == io.EOF
		if !done && err != nil {
			return
		}
		_, err = w.Write(buffer[:n])
		if err != nil {
			log.Println("[xhttp]:\tcould not write response:", err)
		}
	}
}

func writeResource(status int, resource *Resource, w http.ResponseWriter) {
	w.Header().Set("Content-Type", resource.Mime)
	w.WriteHeader(status)

	if resource.Stream != nil {
		done := false
		buffer := make([]byte, 1024)
		for !done {
			n, err := resource.Stream.Read(buffer)
			done = err == io.EOF
			if !done && err != nil {
				return
			}
			_, err = w.Write(buffer[:n])
			if err != nil {
				log.Println("[xhttp]:\tcould not write response:", err)
			}
		}
	} else if resource.BytesData != nil {
		_, _ = w.Write(resource.BytesData)
	}
}

func writeRedirect(w http.ResponseWriter, red *RedirectURL) {
	if red.Params != nil && len(red.Params) > 0 {
		values := url.Values{}
		for n, v := range red.Params {
			values.Add(n, v)
		}
		red.URL = fmt.Sprintf("%s?%s", red.URL, values.Encode())
	}

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
