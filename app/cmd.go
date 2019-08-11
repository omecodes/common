package app

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	registrypb "github.com/zoenion/common/proto/registry"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/credentials"

	"github.com/zoenion/common/prompt"
)

type Vars struct {
	Name         string
	Dir          string
	Domain       string
	IP           string
	ConfigServer string

	Registry       string
	Namespace      string
	RegistryID     string
	registryClient registrypb.RegistryClient

	GRPCAuthorityAddress        string
	AuthorityCertPath           string
	AuthorityCredentials        string
	authorityCert               *x509.Certificate
	authorityGRPCAuthentication credentials.PerRPCCredentials

	GatewayGRPCPort string
	GatewayHTTPPort string

	gRPCAuthorityCredentials credentials.TransportCredentials
	serviceCert              *x509.Certificate
	serviceKey               crypto.PrivateKey
}

type ConfigVars struct {
	Dir  string
	Name string
}

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
	configureCMD.PersistentFlags().StringVar(&cVars.Name, "name", "", "Unique name in registry group")
	configureCMD.PersistentFlags().StringVar(&cVars.Dir, "dir", "", "Configs directory path")

	startCMD := &cobra.Command{
		Use:   "start",
		Short: "start node",
		Run: func(cmd *cobra.Command, args []string) {
			err := StartNode(node, vars)
			if err != nil {
				log.Fatalf("could not start node %s: %s\n", vars.Name, err)
			}
			defer StopNode(node, vars)
			<-prompt.QuitSignal()
		},
	}
	startCMD.PersistentFlags().StringVar(&vars.Name, "name", "", "Unique name in registry group")
	startCMD.PersistentFlags().StringVar(&vars.Dir, "dir", "", "Configs directory path")
	startCMD.PersistentFlags().StringVar(&vars.GatewayGRPCPort, "grpc", "", "GRPC listen address")
	startCMD.PersistentFlags().StringVar(&vars.GatewayHTTPPort, "http", "", "HTTP listen address")
	startCMD.PersistentFlags().StringVar(&vars.Registry, "registry", "", "ArgRegistry location")
	startCMD.PersistentFlags().StringVar(&vars.ConfigServer, "cfg-server", "", "Config server location")
	startCMD.PersistentFlags().StringVar(&vars.IP, "ip", "", "Network address to bind to")
	startCMD.PersistentFlags().StringVar(&vars.Domain, "domain", "", "ArgDomain name")
	startCMD.PersistentFlags().StringVar(&vars.Namespace, "namespace", "", "Group identifier for registry")
	startCMD.PersistentFlags().StringVar(&vars.AuthorityCertPath, "authority-cert", "", "Authority certificate path")
	startCMD.PersistentFlags().StringVar(&vars.GRPCAuthorityAddress, "authority", "", "Authority address location")
	startCMD.PersistentFlags().StringVar(&vars.AuthorityCredentials, "authority-cred", "", "Authority access credentials")

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
		return errors.New("flag --name must be passed")
	}

	if vars.Domain == "" && vars.IP == "" {
		return errors.New("one or both --domain and --ip flags must be passed")
	}

	if vars.Dir == "" {
		d := getDir()
		vars.Dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("could not create %s. Might not be writeable\n", vars.Dir)
			return err
		}
	} else {
		var err error
		vars.Dir, err = filepath.Abs(vars.Dir)
		if err != nil {
			log.Printf("could not find %s\n", vars.Dir)
			return err
		}
	}

	if vars.GRPCAuthorityAddress != "" || vars.AuthorityCertPath != "" || vars.AuthorityCredentials != "" {
		if vars.GRPCAuthorityAddress == "" || vars.AuthorityCertPath == "" || vars.AuthorityCredentials == "" {
			return fmt.Errorf("to enable connection to authority %s, %s and %s flags must me passed", vars.GRPCAuthorityAddress, vars.AuthorityCertPath, vars.AuthorityCredentials)
		}
	}

	if vars.Registry != "" || vars.Namespace != "" {
		if vars.Registry == "" || vars.Namespace == "" {
			return errors.New("to enable connection to registry both --registry and --namespace flags must me passed")
		}
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
