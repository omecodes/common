package app

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"github.com/zoenion/common/app/lang"
	"github.com/zoenion/common/app/templates"
	"github.com/zoenion/common/app/web"
	"github.com/zoenion/common/conf"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type App struct {
	sync.Mutex
	vendor string
	name   string

	options     *options
	initialized bool

	cmd          *cobra.Command
	configureCMD *cobra.Command
	startCMD     *cobra.Command
	versionCMD   *cobra.Command

	translationsDir string
	templatesDir    string
	dataDir         string
	cacheDir        string

	Resources *Resources
	configs   conf.Map
}

func (a *App) Initialize() error {
	return a.init()
}

func (a *App) init() error {
	if a.initialized {
		return nil
	}
	a.initialized = true

	a.Lock()
	defer a.Unlock()

	execName := a.name
	if a.options.instanceName != "" {
		execName = a.options.instanceName
	}
	a.cmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: fmt.Sprintf("Run %s help command", execName),
		Run: func(cmd *cobra.Command, args []string) {
			if a.options.startCMDFunc != nil {
				err := a.initDirs()
				if err != nil {
					log.Fatalln(err)
				}

				err = a.initResources()
				if err != nil {
					log.Fatalln(err)
				}

				if len(a.options.configItems) > 0 {
					cfgFilename := filepath.Join(a.dataDir, "configs.json")
					err = conf.Load(cfgFilename, &a.configs)
					if err != nil {
						log.Fatalln(err)
					}
				}

				a.options.startCMDFunc()
			} else {
				if err := cmd.Help(); err != nil {
					log.Fatalln(err)
				}
			}
		},
	}

	// add configure command
	if len(a.options.configItems) > 0 {
		a.configureCMD = &cobra.Command{
			Use:   "configure",
			Short: fmt.Sprintf("Configure %s", execName),
			Run: func(cmd *cobra.Command, args []string) {
				err := a.initDirs()
				if err != nil {
					log.Fatalln(err)
				}

				configFilename := filepath.Join(a.dataDir, "configs.json")
				oldConf := conf.Map{}
				err = conf.Load(configFilename, &oldConf)

				err = a.configure(configFilename, os.ModePerm, a.options.configItems...)
				if err != nil {
					log.Fatalln(err)
				}

				if a.options.afterConfigure != nil {
					err = a.options.afterConfigure(a.configs, configFilename)
					if err != nil {
						log.Fatalln(err)
					}
				}
			},
		}
		a.cmd.AddCommand(a.configureCMD)
	}

	// add run command
	if a.options.startCMDFunc != nil {
		a.startCMD = &cobra.Command{
			Use:   "start",
			Short: fmt.Sprintf("Start %s", execName),
			Run: func(cmd *cobra.Command, args []string) {
				err := a.initDirs()
				if err != nil {
					log.Fatalln(err)
				}

				err = a.initResources()
				if err != nil {
					log.Fatalln(err)
				}

				cfgFilename := filepath.Join(a.dataDir, "configs.json")
				err = conf.Load(cfgFilename, &a.configs)
				if err != nil {
					log.Fatalln(err)
				}
				a.options.startCMDFunc()
			},
		}
		a.cmd.AddCommand(a.startCMD)
	}

	// add version command
	if a.options.version != "" {
		a.versionCMD = &cobra.Command{
			Use:   "version",
			Short: "Displays app name and version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(a.options.version)
			},
		}
		a.cmd.AddCommand(a.versionCMD)
	}

	return nil
}

func (a *App) initDirs() error {
	// initializing directories
	dirs := configdir.New(a.vendor, a.name)
	globalFolder := dirs.QueryFolders(configdir.Global)[0]
	cacheFolder := dirs.QueryFolders(configdir.Cache)[0]

	if a.options.instanceName != "" {
		a.dataDir = filepath.Join(globalFolder.Path, a.options.instanceName)
	} else {
		a.dataDir = globalFolder.Path
	}

	if a.options.version != "" {
		a.dataDir = filepath.Join(a.dataDir, fmt.Sprintf("v%s", a.options.version))
	}

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

func (a *App) initResources() error {
	// initializing resources manager is required
	if a.options.withResources {
		a.Resources = new(Resources)

		webDir := filepath.Join(a.dataDir, "res", "www")
		err := os.MkdirAll(webDir, os.ModePerm)
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

func (a *App) configure(outputFilename string, mode os.FileMode, items ...configItem) error {
	oldValues := conf.Map{}
	_ = conf.Load(outputFilename, &oldValues)

	newValues := conf.Map{}
	for _, item := range items {
		key := item.configType.String()
		itemOldValues := oldValues.GetConf(key)

		values, err := item.create(item.description, itemOldValues)
		if err != nil {
			return err
		}
		newValues.Set(key, values)
	}
	return newValues.Save(outputFilename, mode)
}

func (a *App) GetConfig(item ConfigType) conf.Map {
	return a.configs.GetConf(item.String())
}

func (a *App) SetName(name string) {
	a.name = name
}

func (a *App) GetCommand() *cobra.Command {
	return a.cmd
}

func (a *App) StartCommand() *cobra.Command {
	return a.startCMD
}

func (a *App) ConfigureCommand() *cobra.Command {
	return a.configureCMD
}

func (a *App) InitDirs() error {
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

func (a *App) LoadConfigs() error {
	cfgFilename := filepath.Join(a.dataDir, "configs.json")
	return conf.Load(cfgFilename, &a.configs)
}

func (a *App) DataDir() string {
	return a.dataDir
}

func (a *App) CacheDir() string {
	return a.cacheDir
}

func (a *App) Label() string {
	return strcase.ToDelimited(a.name, ' ')
}

func (a *App) Name() string {
	return a.name
}

func New(vendor string, name string, opts ...Option) *App {
	a := &App{
		vendor:  vendor,
		name:    name,
		options: new(options),
		configs: conf.Map{},
	}
	for _, opt := range opts {
		opt(a.options)
	}
	return a
}
