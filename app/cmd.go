package app

import "github.com/spf13/cobra"

var (
	ArgRegistry      string
	ArgAddressGRPC   string
	ArgAddressHTTP   string
	ArgDir           string
	ArgConfigsServer string
	ArgDomain        string
	ArgIP            string
	ArgNamespace     string
	ArgName          string
)

type RunFunc func()

type SubCommand interface {
	Name() string
	Short() string
	Run()
}

func AppCommand(use string, subCommands ...SubCommand) *cobra.Command {
	var CMD = &cobra.Command{
		Use:   use,
		Short: use,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	CMD.PersistentFlags().StringVar(&ArgDir, "dir", "", "Configs directory path")
	CMD.PersistentFlags().StringVar(&ArgAddressGRPC, "grpc", "", "GRPC listen address")
	CMD.PersistentFlags().StringVar(&ArgAddressHTTP, "http", "", "HTTP listen address")
	CMD.PersistentFlags().StringVar(&ArgRegistry, "registry", "", "ArgRegistry location")
	CMD.PersistentFlags().StringVar(&ArgConfigsServer, "cfg-server", "", "Config server location")
	CMD.PersistentFlags().StringVar(&ArgIP, "ip", "", "Network address to bind to")
	CMD.PersistentFlags().StringVar(&ArgDomain, "domain", "", "ArgDomain name")
	CMD.PersistentFlags().StringVar(&ArgNamespace, "namespace", "", "Group identifier for registry")
	CMD.PersistentFlags().StringVar(&ArgName, "name", "", "Unique name in registry group")

	for _, sc := range subCommands {
		CMD.AddCommand(&cobra.Command{
			Use:   sc.Name(),
			Short: sc.Short(),
			Run: func(cmd *cobra.Command, args []string) {
				sc.Run()
			},
		})
	}
	return CMD
}

func ConfigureCMD(run func()) SubCommand {
	return &subCMD{
		name:  "configure",
		short: "configure",
		run:   run,
	}
}

func StartCMD(run func()) SubCommand {
	return &subCMD{
		name:  "start",
		short: "start",
		run:   run,
	}
}

type subCMD struct {
	name, short string
	run         func()
}

func (s *subCMD) Name() string {
	return s.name
}
func (s *subCMD) Short() string {
	return s.short
}
func (s *subCMD) Run() {
	s.run()
}
