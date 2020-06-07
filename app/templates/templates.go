package templates

import (
	"github.com/omecodes/common/futils"
	htmpl "html/template"
	"io"
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
	contentType := futils.Mime(resPath)

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
