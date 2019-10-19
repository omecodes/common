package templates

import (
	"bytes"
	htmpl "html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

func (r *Templates) Load(locale string, name string, data interface{}) (string, []byte, error) {
	resPath := filepath.Join(r.dir, locale, name)

	f, err := os.Open(resPath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	contentType := "text/plain"
	contentType, err = fileContentType(f)
	if err != nil {
		panic(err)
	}

	buff := bytes.NewBuffer(nil)
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", nil, err
	}
	switch contentType {
	case "text/html":
		ht := htmpl.New("ht")
		tpl, err := ht.New("html-template").Parse(string(content))
		if err != nil {
			return contentType, nil, err
		}
		err = tpl.Execute(buff, &data)
	default:
		tt := template.New("tt")
		tpl, err := tt.New("text-template").Parse(string(content))
		if err != nil {
			return contentType, nil, err
		}
		err = tpl.Execute(buff, &data)
	}

	return contentType, buff.Bytes(), err
}

func fileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
