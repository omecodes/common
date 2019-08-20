package service

import (
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
	if v.authorityGRPC != "" {
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

	if v.registryAddress != "" {
		if ai := node.Info(); ai != nil {
			ai.Namespace = v.namespace
			registryID, err := v.registry.Register(ai)
			if err != nil {
				log.Printf("could not register %s: %s\n", ai.Name, err)
			} else {
				log.Printf("%s registered as %s", v.name, registryID)
				v.registryID = registryID
			}
		}
	}
	return nil
}

func StopNode(node Node, v *Vars) {
	defer node.Stop()
	if v.registryAddress != "" {
		if err := v.registry.Deregister(v.registryID); err != nil {
			log.Printf("could not deregister from registryAddress: %s\n", err)
		}
	}
}
