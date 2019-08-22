package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoenion/common/auth"
	crypto2 "github.com/zoenion/common/crypto"
	"github.com/zoenion/common/prompt"
	"google.golang.org/grpc/credentials"
)

var (
	Vendor = "Zoenion"
	Name   = strings.Split(os.Args[0], fmt.Sprintf("%c", os.PathSeparator))[0]
)

func CMD(use string, node Node) *cobra.Command {
	vars := new(Vars)
	cVars := new(ConfigVars)

	configureCMD := &cobra.Command{
		Use:   "configure",
		Short: "configure node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateConfVars(cVars); err != nil {
				log.Fatalln(err)
			}
			err := node.Configure(cVars)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	configureCMD.PersistentFlags().StringVar(&cVars.name, CmdFlagName, "", "Unique name in registryAddress group")
	configureCMD.PersistentFlags().StringVar(&cVars.dir, CmdFlagDir, "", "Configs directory path")

	startCMD := &cobra.Command{
		Use:   "start",
		Short: "start node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateRunVars(vars); err != nil {
				log.Fatalln(err)
			}

			if err := loadTools(vars); err != nil {
				log.Fatalln(err)
			}

			defer StopNode(node, vars)

			if err := StartNode(node, vars); err != nil {
				log.Fatalf("starting %s service: %s\n", vars.Name, err)
			}

			<-prompt.QuitSignal()
		},
	}
	startCMD.PersistentFlags().StringVar(&vars.name, CmdFlagName, "", "Unique name in registryAddress group")
	startCMD.PersistentFlags().StringVar(&vars.name, CmdFlagDir, "", "Configs directory path")
	startCMD.PersistentFlags().StringVar(&vars.certificatePath, CmdFlagCert, "", "Public certificate path")
	startCMD.PersistentFlags().StringVar(&vars.keyPath, CmdFlagKey, "", "Private key path")

	startCMD.PersistentFlags().StringVar(&vars.gatewayGRPCPort, CmdFlagGRPC, "", "GRPC Port: gRPC port")
	startCMD.PersistentFlags().StringVar(&vars.gatewayHTTPPort, CmdFlagHTTP, "", "HTTP Port: HTTP port")
	startCMD.PersistentFlags().StringVar(&vars.registryAddress, CmdFlagRegistry, "", "Registry Server - address location")
	startCMD.PersistentFlags().BoolVar(&vars.registrySecure, CmdFlagRegistrySecure, false, "Registry Secure Mode - enable secure connection to registry")
	startCMD.PersistentFlags().StringVar(&vars.namespace, CmdFlagNamespace, "", "Namespace - Group identifier for registryAddress")
	startCMD.PersistentFlags().StringVar(&vars.ip, CmdFlagIP, "", "Network - ip address to listen to. Must matching domain if provided")
	startCMD.PersistentFlags().StringVar(&vars.domain, CmdFlagDomain, "", "Domain - Domain name to bind to")
	startCMD.PersistentFlags().StringVar(&vars.authorityCertPath, CmdFlagAuthorityCert, "", "Authority Certificate - file path")
	startCMD.PersistentFlags().StringVar(&vars.authorityGRPC, CmdFlagAuthority, "", "Authority GRPC - address location")
	startCMD.PersistentFlags().StringVar(&vars.authorityCredentials, CmdFlagAuthorityCred, "", "Authority Credentials - authority authentication credentials")

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

func validateRunVars(vars *Vars) error {
	if vars.name == "" {
		return errors.New("command line: --name flags is required")
	}

	if vars.domain == "" && vars.ip == "" {
		return errors.New("command line: one or both --domain and --ip flags must be passed")
	}

	if vars.dir == "" {
		d := getDir()
		vars.dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("command line: could not create %s. Might not be writeable\n", vars.Dir)
			return err
		}
	} else {
		var err error
		vars.dir, err = filepath.Abs(vars.dir)
		if err != nil {
			log.Printf("command line: could not find %s\n", vars.Dir)
			return err
		}
	}

	if vars.authorityGRPC != "" || vars.authorityCertPath != "" || vars.authorityCredentials != "" {
		if vars.authorityGRPC == "" || vars.authorityCertPath == "" || vars.authorityCredentials == "" {
			return fmt.Errorf("command line: --a-grpc must always be provided with --a-cert and --a-cred")
		}
	}

	if vars.registryAddress != "" || vars.namespace != "" {
		if vars.registryAddress == "" || vars.namespace == "" {
			return errors.New("command line: --namespace must always be provided with --registryAddress")
		}
	}

	if vars.certificatePath != "" || vars.keyPath != "" {
		if vars.certificatePath == "" || vars.keyPath == "" {
			return errors.New("command line: --cert must always be provided with --key")
		}
	}

	return nil
}

func loadTools(v *Vars) error {
	var err error

	if v.certificatePath != "" {
		v.loaded.serviceCert, err = crypto2.LoadCertificate(v.certificatePath)
		if err != nil {
			return fmt.Errorf("could not load service certificate: %s", err)
		}

		v.loaded.serviceKey, err = crypto2.LoadPrivateKey(nil, v.keyPath)
		if err != nil {
			return fmt.Errorf("could not load service private key: %s", err)
		}
	}

	if v.authorityGRPC != "" {
		v.authorityCert, err = crypto2.LoadCertificate(v.authorityCertPath)
		if err != nil {
			return fmt.Errorf("could not load authority certificate: %s", err)
		}

		v.authorityGRPCTransportCredentials, err = credentials.NewClientTLSFromFile(v.authorityCertPath, "")
		if err != nil {
			return fmt.Errorf("could not create authority client credentials: %s", v.authorityCertPath)
		}

		parts := strings.Split(v.authorityCredentials, ":")
		v.authorityClientAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])
	}

	if v.authorityCertPath != "" {
		err = loadSignedKeyPair(v)
		if err != nil {
			return err
		}
	}

	if v.registryAddress != "" {
		if v.registrySecure {
			v.registry = NewSyncRegistry(v.registryAddress, ClientMutualTLS(v))
		} else {
			v.registry = NewSyncRegistry(v.registryAddress, nil)
		}
	}
	return nil
}

func validateConfVars(args *ConfigVars) error {
	if args.dir == "" {
		d := getDir()
		args.dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("could not create %s. Might not be writeable\n", args.Dir)
			return err
		}

	} else {
		var err error
		args.dir, err = filepath.Abs(args.dir)
		if err != nil {
			log.Printf("could not find %s\n", args.Dir)
			return err
		}
	}
	return nil
}
