package cmd

import "github.com/spf13/cobra"

var (
	Registry      string
	AddressGRPC   string
	AddressHTTP   string
	Dir           string
	ConfigsServer string
	Domain        string
	IP            string
	Namespace     string
	Name          string
)

func Gateway(use string) *cobra.Command {
	var CMD = &cobra.Command{
		Use:   use,
		Short: use,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	CMD.PersistentFlags().StringVar(&Dir, "dir", "", "Configs directory path")
	CMD.PersistentFlags().StringVar(&AddressGRPC, "grpc", "", "GRPC listen address")
	CMD.PersistentFlags().StringVar(&AddressHTTP, "http", "", "HTTP listen address")
	CMD.PersistentFlags().StringVar(&Registry, "registry", "", "Registry location")
	CMD.PersistentFlags().StringVar(&ConfigsServer, "cfg-server", "", "Config server location")
	CMD.PersistentFlags().StringVar(&IP, "ip", "", "Network address to bind to")
	CMD.PersistentFlags().StringVar(&Domain, "domain", "", "Domain name")
	CMD.PersistentFlags().StringVar(&Namespace, "namespace", "", "Group identifier for registry")
	CMD.PersistentFlags().StringVar(&Name, "name", "", "Unique name in registry group")
	return CMD
}
