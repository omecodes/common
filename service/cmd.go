package service

import (
	"fmt"
	"github.com/iancoleman/strcase"
	crypto2 "github.com/zoenion/common/crypto"
	servicepb "github.com/zoenion/common/proto/service"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoenion/common/prompt"
)

var (
	Vendor = "Zoenion"
	Name   = strings.Split(os.Args[0], fmt.Sprintf("%c", os.PathSeparator))[0]
)

func CMD(use string, service Service) *cobra.Command {
	params := BoxParams{}
	var cfgDir, cfgName string

	configureCMD := &cobra.Command{
		Use:   "configure",
		Short: "configure node",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgDir == "" {
				d := getDir()
				if err := d.Create(); err != nil {
					log.Fatalln("could not initialize configs dir:", err)
				}
				cfgDir = d.path
			}

			if err := validateConfVars(cfgName, cfgDir); err != nil {
				log.Fatalln(err)
			}
			err := service.Configure(cfgName, cfgDir)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	configureCMD.PersistentFlags().StringVar(&cfgName, CmdFlagName, "", "Unique name in registryAddress group")
	configureCMD.PersistentFlags().StringVar(&cfgDir, CmdFlagDir, "", "Configs directory path")

	startCMD := &cobra.Command{
		Use:   "start",
		Short: "start node",
		Run: func(cmd *cobra.Command, args []string) {
			box := new(Box)
			if params.Name == "" {
				params.Name = Name
			}

			if params.Dir == "" {
				d := getDir()
				err := d.Create()
				if err != nil {
					log.Fatalf("could not initialize configs dir: %s\n", err)
				}
				params.Dir = d.path
			}

			box.params = params
			if err := box.validateParams(); err != nil {
				log.Fatalln(err)
			}

			if err := box.loadTools(); err != nil {
				log.Fatalln(err)
			}

			bc, err := service.Configs(params.Name, params.Dir)
			if err != nil {
				log.Fatalf("could not load box configs: %s\n", err)
			}

			if err := box.start(bc); err != nil {
				log.Fatalf("starting %s service: %s\n", box.Name, err)
			}
			if box.registry != nil {
				certEncoded, _ := crypto2.PEMEncodeCertificate(box.serviceCert)
				box.params.RegistryID, err = box.registry.Register(&servicepb.Info{
					Name:      strcase.ToDelimited(box.Name(), '-'),
					Namespace: box.params.Namespace,
					Type:      service.Type(),
					Label:     strcase.ToCamel(box.params.Name),
					Nodes:     box.gateway.nodes(),
					Meta:      map[string]string{MetaCertificate: string(certEncoded)},
				})
				if err != nil {
					log.Printf("could not register service: %s\n", err)
				}
			}
			service.AfterStart()

			<-prompt.QuitSignal()

			box.stop()
			if box.params.RegistryID != "" {
				err = box.registry.Deregister(box.params.RegistryID)
				if err != nil {
					log.Printf("could not de-register service: %s\n", err)
				}
			}
			service.AfterStop()
		},
	}
	startCMD.PersistentFlags().StringVar(&params.Name, CmdFlagName, "", "Unique name in registryAddress group")
	startCMD.PersistentFlags().StringVar(&params.Name, CmdFlagDir, "", "Configs directory path")
	startCMD.PersistentFlags().StringVar(&params.CertificatePath, CmdFlagCert, "", "Public certificate path")
	startCMD.PersistentFlags().StringVar(&params.KeyPath, CmdFlagKey, "", "Private key path")
	startCMD.PersistentFlags().StringVar(&params.GatewayGRPCPort, CmdFlagGRPC, "", "Grpc Port: gRPC port")
	startCMD.PersistentFlags().StringVar(&params.GatewayHTTPPort, CmdFlagHTTP, "", "Web Port: Web port")
	startCMD.PersistentFlags().StringVar(&params.RegistryAddress, CmdFlagRegistry, "", "Registry Server - address location")
	startCMD.PersistentFlags().BoolVar(&params.RegistrySecure, CmdFlagRegistrySecure, false, "Registry Secure Mode - enable secure connection to registry")
	startCMD.PersistentFlags().StringVar(&params.Namespace, CmdFlagNamespace, "", "Namespace - Group identifier for registryAddress")
	startCMD.PersistentFlags().StringVar(&params.Ip, CmdFlagIP, "", "Network - ip address to listen to. Must matching domain if provided")
	startCMD.PersistentFlags().StringVar(&params.Domain, CmdFlagDomain, "", "Domain - Domain name to bind to")
	startCMD.PersistentFlags().StringVar(&params.CaCertPath, CmdFlagAuthorityCert, "", "Authority Certificate - file path")
	startCMD.PersistentFlags().StringVar(&params.CaGRPC, CmdFlagAuthority, "", "Authority Grpc - address location")
	startCMD.PersistentFlags().StringVar(&params.CaCredentials, CmdFlagAuthorityCred, "", "Authority Credentials - authority authentication credentials")

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

func validateConfVars(name, dir string) error {
	if dir == "" {
		d := getDir()
		dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("could not create %s. Might not be writeable\n", dir)
			return err
		}

	} else {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			log.Printf("could not find %s\n", dir)
			return err
		}
	}
	return nil
}
