package templates

import (
	htmpl "html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	DefaultLocal       = "fr-FR"
	TmplLoginMail      = ""
	RegistrationEmail  = "registration_email"
	PasswordResetEmail = "reset_password_email"
)

func New(dir string) *Templates {
	return &Templates{dir: dir}
}

type Templates struct {
	dir string
}

func (r *Templates) Load(locale string, name string, data interface{}, out io.Writer) (string, error) {
	resPath := filepath.Join(r.dir, locale, name)
	baseName := filepath.Base(resPath)
	contentType := fileContentType(resPath)

	if strings.HasPrefix(contentType, "text/html") {
		ht := htmpl.New(baseName)
		tpl, err := ht.ParseFiles(resPath)
		if err != nil {
			return contentType, err
		}
		return contentType, tpl.Execute(out, &data)

	} else {
		tt := template.New(baseName)
		tpl, err := tt.ParseFiles(resPath)
		if err != nil {
			return contentType, err
		}
		return contentType, tpl.Execute(out, &data)
	}
}

func fileContentType(filename string) (contentType string) {
	contentType = "text/plain"

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return
	}

	contentType = http.DetectContentType(buffer)
	return
}
