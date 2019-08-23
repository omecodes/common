package service

import (
	"fmt"
	"log"
)

func startBox(service Service, v *Box) error {
	if v.authorityGRPC != "" {
		if err := loadSignedKeyPair(v); err != nil {
			return fmt.Errorf("could not load certificate/key: %s", err)
		}
	}

	if v.registryAddress != "" {
		if ai := service.Info(); ai != nil {
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

func stopBox(node Service, v *Box) {
	defer node.Stop()
	if v.registryAddress != "" {
		if err := v.registry.Deregister(v.registryID); err != nil {
			log.Printf("could not deregister from registryAddress: %s\n", err)
		}
	}
}
