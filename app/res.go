package app

import (
	"github.com/omecodes/common/app/lang"
	"github.com/omecodes/common/app/templates"
	"github.com/omecodes/common/app/web"
	"golang.org/x/text/language"
	"io"
)

type Resources struct {
	dir       string
	i18n      *lang.I18n
	templates *templates.Templates
	web       *web.Server
}

func (r *Resources) ResolveLanguage(acceptLanguageHeader string) language.Tag {
	return r.i18n.LanguageFromAcceptLanguageHeader(acceptLanguageHeader)
}

func (r *Resources) Translated(locale language.Tag, name string, args ...interface{}) string {
	return r.i18n.Translated(locale, name, args...)
}

func (r *Resources) LoadTemplate(locale language.Tag, name string, data interface{}, out io.Writer) (string, error) {
	return r.templates.Load(locale.String(), name, data, out)
}

func (r *Resources) FileContent(locale language.Tag, name string) (string, io.ReadCloser, int64, error) {
	return r.web.Serve(locale.String(), name)
}

func (r *Resources) Translations(locale language.Tag) (map[string]string, error) {
	return r.i18n.Translations(locale)
}
