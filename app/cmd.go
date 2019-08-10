package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zoenion/common/prompt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type RunArgs struct {
	Dir               string
	Registry          string
	AddressGRPC       string
	AddressHTTP       string
	ConfigsServer     string
	Domain            string
	IP                string
	Namespace         string
	Name              string
	AuthorityCertPath string
}

type ConfigArgs struct {
	Dir  string
	Name string
}

var (
	Vendor = "Zoenion"
	Name   = strings.Split(os.Args[0], fmt.Sprintf("%c", os.PathSeparator))[0]
)

func CMD(use string, node Node) *cobra.Command {

	startArgs := new(RunArgs)
	cfgArgs := new(ConfigArgs)

	configureCMD := &cobra.Command{
		Use:   "configure",
		Short: "configure node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateConfArgs(cfgArgs); err != nil {
				log.Fatalln(err)
			}
			err := node.Configure(cfgArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	configureCMD.PersistentFlags().StringVar(&cfgArgs.Name, "name", "", "Unique name in registry group")
	configureCMD.PersistentFlags().StringVar(&cfgArgs.Dir, "dir", "", "Configs directory path")

	startCMD := &cobra.Command{
		Use:   "start",
		Short: "start node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateStartArgs(startArgs); err != nil {
				log.Fatalln(err)
			}
			if err := node.Init(startArgs); err != nil {
				log.Fatalln(err)
			}
			if err := node.Start(); err != nil {
				log.Fatalln(err)
			}
			defer node.Stop()
			<-prompt.QuitSignal()
		},
	}
	startCMD.PersistentFlags().StringVar(&startArgs.Name, "name", "", "Unique name in registry group")
	startCMD.PersistentFlags().StringVar(&startArgs.Dir, "dir", "", "Configs directory path")
	startCMD.PersistentFlags().StringVar(&startArgs.AddressGRPC, "grpc", "", "GRPC listen address")
	startCMD.PersistentFlags().StringVar(&startArgs.AddressHTTP, "http", "", "HTTP listen address")
	startCMD.PersistentFlags().StringVar(&startArgs.Registry, "registry", "", "ArgRegistry location")
	startCMD.PersistentFlags().StringVar(&startArgs.ConfigsServer, "cfg-server", "", "Config server location")
	startCMD.PersistentFlags().StringVar(&startArgs.IP, "ip", "", "Network address to bind to")
	startCMD.PersistentFlags().StringVar(&startArgs.Domain, "domain", "", "ArgDomain name")
	startCMD.PersistentFlags().StringVar(&startArgs.Namespace, "namespace", "", "Group identifier for registry")
	startCMD.PersistentFlags().StringVar(&startArgs.AuthorityCertPath, "authority-cert", "", "Authority certificate path")

	command := &cobra.Command{
		Use:   use,
		Short: use,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	command.AddCommand(configureCMD)
	command.AddCommand(startCMD)
	return command
}

func validateStartArgs(args *RunArgs) error {
	if args.Dir == "" {
		d := getDir()
		args.Dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("could not create %s. Might not be writeable\n", args.Dir)
			return err
		}

	} else {
		var err error
		args.Dir, err = filepath.Abs(args.Dir)
		if err != nil {
			log.Printf("could not find %s\n", args.Dir)
			return err
		}
	}
	return nil
}

func validateConfArgs(args *ConfigArgs) error {
	if args.Dir == "" {
		d := getDir()
		args.Dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("could not create %s. Might not be writeable\n", args.Dir)
			return err
		}

	} else {
		var err error
		args.Dir, err = filepath.Abs(args.Dir)
		if err != nil {
			log.Printf("could not find %s\n", args.Dir)
			return err
		}
	}
	return nil
}
