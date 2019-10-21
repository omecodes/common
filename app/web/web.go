package web

import (
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/futils"
	"io"
	"os"
	"path/filepath"
)

var mimes = map[string]string{
	".js":   "text/javascript",
	".html": "text/html",
	".csh":  "text/x-script.csh",
	".css":  "text/css",
}

func New(dir string) *Server {
	return &Server{dir: dir}
}

type Server struct {
	dir string
}

func (s *Server) Serve(locale string, filename string) (string, io.ReadCloser, int64, error) {
	if locale == "" {
		locale = "en-US"
	}
	localeFolder := filepath.Join(s.dir, locale)
	if !futils.FileExists(localeFolder) {
		locale = "en-US"
	}

	localeFolder = filepath.Join(s.dir, locale)
	if !futils.FileExists(localeFolder) {
		return "", nil, 0, errors.NotFound
	}

	resPath := filepath.Join(localeFolder, filename)

	size := int64(0)
	contentType := ""
	extension := filepath.Ext(resPath)
	if extension == "" {
		resPath = filepath.Join(resPath, "index.html")
		extension = "html"
	}

	f, err := os.Open(resPath)
	if err != nil {
		return "", f, size, err
	}

	stat, err := f.Stat()
	if err != nil {
		return "", f, size, err
	}

	size = stat.Size()
	contentType, ok := mimes[extension]
	if !ok {
		extension = futils.Mime(filename)
	}

	return contentType, f, size, err
}
