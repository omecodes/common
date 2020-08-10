package app

import (
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/futils"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/lang"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var mimes = map[string]string{
	".js":   "text/javascript",
	".html": "text/html",
	".csh":  "text/x-script.csh",
	".css":  "text/css",
	".svg":  "image/svg+xml",
}

func NewFolder(dir string, i18n *lang.I18n) *Webapp {
	return &Webapp{dir: dir, i18n: i18n}
}

type Webapp struct {
	dir  string
	i18n *lang.I18n
}

func (s *Webapp) Serve(locale string, appName string, filename string) (string, io.ReadCloser, int64, error) {
	if locale == "" {
		locale = "en-US"
	}

	localeFolder := filepath.Join(s.dir, locale)
	if !futils.FileExists(localeFolder) {
		locale = "en-US"
	}

	localeFolder = filepath.Join(s.dir, locale, appName)
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

	if !futils.FileExists(resPath) {
		return "", nil, 0, errors.NotFound
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

func (s *Webapp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("lang")
	if locale == "" && s.i18n != nil {
		tag := s.i18n.LanguageFromAcceptLanguageHeader(r.Header.Get("Accept-Language"))
		locale = tag.String()
	}

	var appName, filename string
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	appName = parts[0]
	if len(parts) == 2 {
		filename = parts[1]
		if path.Ext(filename) == "" {
			filename = "index.html"
		}
	} else {
		filename = path.Join(parts[1:]...)
	}

	contentType, content, size, err := s.Serve(locale, appName, filename)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	defer content.Close()
	httpx.WriteContent(w, contentType, size, content)
}
