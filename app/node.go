package app

import (
	"context"
	"fmt"
	registrypb "github.com/zoenion/common/proto/registry"
	"log"
)

type Node interface {
	Configure(args *ConfigVars) error
	Init(args *Vars) error
	Start() error
	RegistryInfo() *registrypb.Application
	Stop()
}

func StartNode(node Node, vars *Vars) error {
	ctx := context.Background()

	if err := validateRunVars(vars); err != nil {
		return err
	}

	if vars.GRPCAuthorityAddress != "" {
		if err := loadSignedKeyPair(vars); err != nil {
			return fmt.Errorf("could not load certificate/key: %s", err)
		}
	}

	if err := node.Init(vars); err != nil {
		return err
	}

	if err := node.Start(); err != nil {
		return err
	}

	if vars.Registry != "" {
		if ai := node.RegistryInfo(); ai != nil {
			regClient, err := registryClient(vars)
			if err != nil {
				return fmt.Errorf("could not connect to registry server: %s", err)
			}

			rsp, err := regClient.Register(ctx, &registrypb.RegisterRequest{Application: ai})
			if err != nil {
				log.Printf("could not register %s: %s\n", ai.Name, err)
			} else {
				vars.RegistryID = rsp.Id
				log.Printf("%s registered as %s", vars.Name, rsp.Id)
			}
		}
	}
	return nil
}

func StopNode(node Node, vars *Vars) error {
	if vars.Registry != "" {
		regClient, err := registryClient(vars)
		if err != nil {
			return fmt.Errorf("could not connect to registry server: %s", err)
		}
		_, err = regClient.Deregister(context.Background(), &registrypb.DeregisterRequest{Id: vars.RegistryID})
		return err
	}
	return nil
}
