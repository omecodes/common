package app

import (
	"github.com/shibukawa/configdir"
	"github.com/zoenion/common/app/lang"
	"github.com/zoenion/common/app/templates"
	"github.com/zoenion/common/app/web"
	"log"
	"os"
	"path/filepath"
)

func New(vendor, name, version, label string) *App {
	return &App{
		vendor:  vendor,
		name:    name,
		version: version,
		label:   label,
	}
}

type App struct {
	vendor  string
	name    string
	version string
	label   string

	translationsDir string
	templatesDir    string
	dataDir         string
	cacheDir        string
	//homeDir 		string
	Resources *Resources
}

func (a *App) CreateDirs() error {
	dirs := configdir.New(a.vendor, a.name)
	globalFolder := dirs.QueryFolders(configdir.Global)[0]
	cacheFolder := dirs.QueryFolders(configdir.Cache)[0]

	a.dataDir = globalFolder.Path
	err := os.MkdirAll(a.dataDir, os.ModePerm)
	if err != nil {
		log.Println("could not create configs dir:", err)
		return err
	}

	a.cacheDir = cacheFolder.Path
	err = os.MkdirAll(a.cacheDir, os.ModePerm)
	if err != nil {
		log.Println("could not create cache dir:", err)
		return err
	}
	return nil
}

func (a *App) Init(opts ...Option) error {
	appOptions := new(options)
	for _, opt := range opts {
		opt(appOptions)
	}

	err := a.CreateDirs()
	if err != nil {
		return err
	}

	if appOptions.withResources {
		a.Resources = new(Resources)

		webDir := filepath.Join(a.dataDir, "res", "www")
		err = os.MkdirAll(webDir, os.ModePerm)
		if err != nil {
			log.Println("could not create www dir:", err)
			return err
		}
		a.Resources.web = web.New(webDir)

		templatesDir := filepath.Join(a.dataDir, "res", "templates")
		err = os.MkdirAll(templatesDir, os.ModePerm)
		if err != nil {
			log.Println("could not create templates dir:", err)
			return err
		}

		i18nDir := filepath.Join(a.dataDir, "res", "i18n")
		err = os.MkdirAll(i18nDir, os.ModePerm)
		if err != nil {
			log.Println("could not create i18n dir:", err)
			return err
		}
		a.Resources.templates = templates.New(templatesDir)

		a.Resources.i18n = lang.NewManager(i18nDir)
		err = a.Resources.i18n.Load()
		if err != nil {
			log.Println("could not laod translations")
			return err
		}
	}
	return nil
}

func (a *App) DataDir() string {
	return a.dataDir
}

func (a *App) CacheDir() string {
	return a.cacheDir
}

func (a *App) Label() string {
	return a.label
}
