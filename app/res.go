package app

import (
	"github.com/zoenion/common/app/lang"
	"github.com/zoenion/common/app/templates"
	"golang.org/x/text/language"
	"io"
)

type Resources struct {
	i18n      *lang.I18n
	templates *templates.Templates
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
