package app

import (
	"fmt"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"github.com/zoenion/common/app/lang"
	"github.com/zoenion/common/app/templates"
	"github.com/zoenion/common/app/web"
	"log"
	"os"
	"path/filepath"
)

type CmdRunFunc func()

func New(vendor, name, version, label string) *App {
	return &App{
		vendor:  vendor,
		name:    name,
		version: version,
		label:   label,
	}
}

func WithDefaultCommands(vendor, version, label string, configure, start CmdRunFunc) *App {
	a := &App{
		vendor:  vendor,
		version: version,
		label:   label,
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

			fmt.Printf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n")
			fmt.Printf("%s configs\n", label)
			fmt.Printf("\n\n")
			configure()
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
			start()
		},
	}

	a.versionCMD = &cobra.Command{
		Use:   "version",
		Short: "Displays application name and version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(a.version)
		},
	}

	a.cmd.PersistentFlags().StringVar(&a.name, "name", "", "Instance name. Used as application data folder base name")
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
}

func (a *App) SetName(name string) {
	a.name = name
}

func (a *App) GetCommand() *cobra.Command {
	return a.cmd
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

func (a *App) Name() string {
	return a.name
}
