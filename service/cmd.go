package service

import (
	"errors"
	"fmt"
	"github.com/zoenion/common/auth"
	crypto2 "github.com/zoenion/common/crypto"
	servicepb "github.com/zoenion/common/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	configureCMD.PersistentFlags().StringVar(&cVars.Name, CmdFlagName, "", "Unique name in registry group")
	configureCMD.PersistentFlags().StringVar(&cVars.Dir, CmdFlagDir, "", "Configs directory path")

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
	startCMD.PersistentFlags().StringVar(&vars.Name, CmdFlagName, "", "Unique name in registry group")
	startCMD.PersistentFlags().StringVar(&vars.Dir, CmdFlagDir, "", "Configs directory path")
	startCMD.PersistentFlags().StringVar(&vars.CertificatePath, CmdFlagCert, "", "Public certificate path")
	startCMD.PersistentFlags().StringVar(&vars.KeyPath, CmdFlagKey, "", "Private key path")

	startCMD.PersistentFlags().StringVar(&vars.GatewayGRPCPort, CmdFlagGRPC, "", "GRPC Port: gRPC port")
	startCMD.PersistentFlags().StringVar(&vars.GatewayHTTPPort, CmdFlagHTTP, "", "HTTP Port: HTTP port")
	startCMD.PersistentFlags().StringVar(&vars.Registry, CmdFlagRegistry, "", "Registry Server - address location")
	startCMD.PersistentFlags().StringVar(&vars.RegistryCertPath, CmdFlagRegistryCert, "", "Registry Certificate - server certificate file path")
	startCMD.PersistentFlags().StringVar(&vars.Namespace, CmdFlagNamespace, "", "Namespace - Group identifier for registry")
	startCMD.PersistentFlags().StringVar(&vars.IP, CmdFlagIP, "", "Network - ip address to listen to. Must matching domain if provided")
	startCMD.PersistentFlags().StringVar(&vars.Domain, CmdFlagDomain, "", "Domain - Domain name to bind to")
	startCMD.PersistentFlags().StringVar(&vars.AuthorityCertPath, CmdFlagAuthorityCert, "", "Authority Certificate - file path")
	startCMD.PersistentFlags().StringVar(&vars.AuthorityGRPC, CmdFlagAuthority, "", "Authority GRPC - address location")
	startCMD.PersistentFlags().StringVar(&vars.AuthorityCredentials, CmdFlagAuthorityCred, "", "Authority Credentials - authority authentication credentials")

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
	if vars.Name == "" {
		return errors.New("command line: --name flags is required")
	}

	if vars.Domain == "" && vars.IP == "" {
		return errors.New("command line: one or both --domain and --ip flags must be passed")
	}

	if vars.Dir == "" {
		d := getDir()
		vars.Dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("command line: could not create %s. Might not be writeable\n", vars.Dir)
			return err
		}
	} else {
		var err error
		vars.Dir, err = filepath.Abs(vars.Dir)
		if err != nil {
			log.Printf("command line: could not find %s\n", vars.Dir)
			return err
		}
	}

	if vars.AuthorityGRPC != "" || vars.AuthorityCertPath != "" || vars.AuthorityCredentials != "" {
		if vars.AuthorityGRPC == "" || vars.AuthorityCertPath == "" || vars.AuthorityCredentials == "" {
			return fmt.Errorf("command line: --a-grpc must always be provided with --a-cert and --a-cred")
		}
	}

	if vars.Registry != "" || vars.Namespace != "" {
		if vars.Registry == "" || vars.Namespace == "" {
			return errors.New("command line: --namespace must always be provided with --registry")
		}
	}

	if vars.CertificatePath != "" || vars.KeyPath != "" {
		if vars.CertificatePath == "" || vars.KeyPath == "" {
			return errors.New("command line: --cert must always be provided with --key")
		}
	}

	return nil
}

func loadTools(v *Vars) error {
	var err error

	if v.CertificatePath != "" {
		v.loaded.serviceCert, err = crypto2.LoadCertificate(v.CertificatePath)
		if err != nil {
			return fmt.Errorf("could not load service certificate: %s", err)
		}

		v.loaded.serviceKey, err = crypto2.LoadPrivateKey(nil, v.KeyPath)
		if err != nil {
			return fmt.Errorf("could not load service private key: %s", err)
		}
	}

	if v.Registry != "" {
		if v.RegistryCertPath != "" {
			cred, err := credentials.NewClientTLSFromFile(v.AuthorityCertPath, "")
			if err != nil {
				return fmt.Errorf("could not load registry transpost credentials")
			}
			v.loaded.registryClient, err = servicepb.NewClient(v.Registry, grpc.WithTransportCredentials(cred))
		} else {
			v.loaded.registryClient, err = servicepb.NewClient(v.Registry, grpc.WithInsecure())
		}
		if err != nil {
			return err
		}
	}

	if v.AuthorityGRPC != "" {
		v.loaded.authorityCert, err = crypto2.LoadCertificate(v.AuthorityCertPath)
		if err != nil {
			return fmt.Errorf("could not load authority certificate: %s", err)
		}

		v.loaded.authorityGRPCTransportCredentials, err = credentials.NewClientTLSFromFile(v.AuthorityCertPath, "")
		if err != nil {
			return fmt.Errorf("could not create authority client credentials: %s", v.AuthorityCertPath)
		}

		parts := strings.Split(v.AuthorityCredentials, ":")
		v.loaded.authorityClientAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])
	}

	return nil
}

func validateConfVars(args *ConfigVars) error {
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
