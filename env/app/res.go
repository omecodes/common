package app

import (
	"github.com/omecodes/common/env/web/app"
	templates2 "github.com/omecodes/common/env/web/templates"
	"github.com/omecodes/common/utils/lang"
	"golang.org/x/text/language"
	"io"
)

type Resources struct {
	dir        string
	staticsDir string
	i18n       *lang.I18n
	templates  *templates2.Templates
	web        *app.Webapp
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

func (r *Resources) WAFileContent(locale language.Tag, appName string, filename string) (string, io.ReadCloser, int64, error) {
	return r.web.Serve(locale.String(), appName, filename)
}

func (r *Resources) Translations(locale language.Tag) (map[string]string, error) {
	return r.i18n.Translations(locale)
}

func (r *Resources) StaticDir() string {
	return r.staticsDir
}
