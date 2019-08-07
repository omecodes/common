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

func CMD(use string) *cobra.Command {
	command := &cobra.Command{
		Use:   use,
		Short: use,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	command.PersistentFlags().StringVar(&ArgDir, "dir", "", "Configs directory path")
	command.PersistentFlags().StringVar(&ArgAddressGRPC, "grpc", "", "GRPC listen address")
	command.PersistentFlags().StringVar(&ArgAddressHTTP, "http", "", "HTTP listen address")
	command.PersistentFlags().StringVar(&ArgRegistry, "registry", "", "ArgRegistry location")
	command.PersistentFlags().StringVar(&ArgConfigsServer, "cfg-server", "", "Config server location")
	command.PersistentFlags().StringVar(&ArgIP, "ip", "", "Network address to bind to")
	command.PersistentFlags().StringVar(&ArgDomain, "domain", "", "ArgDomain name")
	command.PersistentFlags().StringVar(&ArgNamespace, "namespace", "", "Group identifier for registry")
	command.PersistentFlags().StringVar(&ArgName, "name", "", "Unique name in registry group")

	return command
}
