package cmd

import (
	"github.com/spf13/cobra"
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
	return CMD
}
