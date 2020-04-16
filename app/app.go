package app

import (
	"fmt"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"github.com/zoenion/common/app/lang"
	"github.com/zoenion/common/app/templates"
	"github.com/zoenion/common/app/web"
	"github.com/zoenion/common/conf"
	"log"
	"os"
	"path/filepath"
)

type RunCommandFunc func()

func New(vendor, version, label string, name string) *App {
	a := &App{
		vendor:  vendor,
		version: version,
		label:   label,
		name:    name,
		configs: conf.Map{},
	}

	err := a.init(false)
	if err != nil {
		log.Fatalln(err)
	}
	return a
}

func WithDefaultCommands(vendor, version, label string, runFunc RunCommandFunc, opts ...Option) *App {
	a := &App{
		vendor:  vendor,
		version: version,
		label:   label,
		configs: conf.Map{},
	}

	var options options
	for _, opt := range opts {
		opt(&options)
	}

	a.cmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: fmt.Sprintf("Configure and run instances of %s", label),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	a.configureCMD = &cobra.Command{
		Use:   "configure",
		Short: fmt.Sprintf("Configure an instance of %s", label),
		Run: func(cmd *cobra.Command, args []string) {
			if a.name == "" {
				if err := cmd.Help(); err != nil {
					log.Fatalln(err)
				}
				log.Fatalln("missing --name flag")
			}

			err := a.init(options.withResources)
			if err != nil {
				log.Fatalln(err)
			}

			configFilename := filepath.Join(a.dataDir, "configs.json")
			oldConf := conf.Map{}
			err = conf.Load(configFilename, &oldConf)

			err = a.configure(configFilename, os.ModePerm, options.configItems...)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	a.startCMD = &cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("Start an instance of %s", label),
		Run: func(cmd *cobra.Command, args []string) {
			if a.name == "" {
				if err := cmd.Help(); err != nil {
					log.Fatalln(err)
				}
				log.Fatalln("missing --name flag")
			}

			err := a.init(options.withResources)
			if err != nil {
				log.Fatalln(err)
			}

			cfgFilename := filepath.Join(a.dataDir, "configs.json")
			err = conf.Load(cfgFilename, &a.configs)
			if err != nil {
				log.Fatalln(err)
			}

			runFunc()
		},
	}
	a.versionCMD = &cobra.Command{
		Use:   "version",
		Short: "Displays app name and version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(a.version)
		},
	}

	a.cmd.PersistentFlags().StringVar(&a.name, "name", "", "Instance name. Used as app data folder base name")
	_ = cobra.MarkFlagRequired(a.cmd.PersistentFlags(), "name")

	a.cmd.AddCommand(a.configureCMD)
	a.cmd.AddCommand(a.startCMD)
	a.cmd.AddCommand(a.versionCMD)

	return a
}

type App struct {
	vendor  string
	name    string
	version string
	label   string

	cmd          *cobra.Command
	configureCMD *cobra.Command
	startCMD     *cobra.Command
	versionCMD   *cobra.Command

	translationsDir string
	templatesDir    string
	dataDir         string
	cacheDir        string
	//homeDir 		string
	Resources *Resources

	configs conf.Map
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

func (a *App) init(withResources bool) error {
	err := a.InitDirs()
	if err != nil {
		return err
	}

	if withResources {
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

func (a *App) Name() string {
	return a.name
}
