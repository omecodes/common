package front

import (
	"github.com/omecodes/common/lang"
	"net/http"
)

type Options struct {
	webAppsFolder          string
	staticFolder           string
	tls                    bool
	apiHandler             http.Handler
	certFilename           string
	keyFilename            string
	webAppsRoutePrefix     string
	staticFilesRoutePrefix string
	apiRoutePrefix         string
	i18n                   *lang.I18n
}

type Option func(opts *Options)

func Translations(n *lang.I18n) Option {
	return func(opts *Options) {
		opts.i18n = n
	}
}

func WebAppsFolder(folder string) Option {
	return func(opts *Options) {
		opts.webAppsFolder = folder
	}
}

func StaticFilesFolder(folder string) Option {
	return func(opts *Options) {
		opts.staticFolder = folder
	}
}

func TLS(certFilename, keyFilename string) Option {
	return func(opts *Options) {
		opts.certFilename = certFilename
		opts.keyFilename = keyFilename
		opts.tls = keyFilename != "" && certFilename != ""
	}
}

func WebAppsRoutePrefix(prefix string) Option {
	return func(opts *Options) {
		opts.webAppsRoutePrefix = prefix
	}
}

func StaticFilesRoutePrefix(prefix string) Option {
	return func(opts *Options) {
		opts.staticFilesRoutePrefix = prefix
	}
}

func ApiRoutePrefix(prefix string) Option {
	return func(opts *Options) {
		opts.apiRoutePrefix = prefix
	}
}
