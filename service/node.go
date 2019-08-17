package service

import (
	"context"
	"fmt"
	servicepb "github.com/zoenion/common/proto/service"
	"log"
)

type Node interface {
	Configure(cVars *ConfigVars) error
	Init(vars *Vars) error
	Start() error
	Info() *servicepb.Info
	Stop()
}

func StartNode(node Node, v *Vars) error {
	ctx := context.Background()

	if v.AuthorityGRPC != "" {
		if err := loadSignedKeyPair(v); err != nil {
			return fmt.Errorf("could not load certificate/key: %s", err)
		}
	}

	if err := node.Init(v); err != nil {
		return err
	}

	if err := node.Start(); err != nil {
		return err
	}

	if v.Registry != "" {
		if ai := node.Info(); ai != nil {
			rsp, err := v.loaded.registryClient.Register(ctx, &servicepb.RegisterRequest{Service: ai})
			if err != nil {
				log.Printf("could not register %s: %s\n", ai.Name, err)
			} else {
				v.RegistryID = rsp.RegistryId
				log.Printf("%s registered as %s", v.Name, rsp.RegistryId)
			}
		}
	}
	return nil
}

func StopNode(node Node, v *Vars) {
	defer node.Stop()
	if v.Registry != "" {
		if _, err := v.loaded.registryClient.Deregister(context.Background(), &servicepb.DeregisterRequest{RegistryId: v.RegistryID}); err != nil {
			log.Printf("could not deregister from registry: %s\n", err)
		}
	}
}
