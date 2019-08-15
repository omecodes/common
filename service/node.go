package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zoenion/common/conf"
	configpb "github.com/zoenion/common/proto/config"
	servicepb "github.com/zoenion/common/proto/service"
	"github.com/zoenion/common/service/net"
	"log"
	"time"
)

type Node interface {
	Configure(cVars *ConfigVars) error
	Init(vars *Vars) error
	Start() error
	Info() *servicepb.Info
	ShareState(chan conf.Map)
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
		if ai := node.Info(); ai != nil {
			regClient, err := registryClient(vars)
			if err != nil {
				return fmt.Errorf("could not connect to registry server: %s", err)
			}

			rsp, err := regClient.Register(ctx, &servicepb.RegisterRequest{Service: ai})
			if err != nil {
				log.Printf("could not register %s: %s\n", ai.Name, err)
			} else {
				vars.RegistryID = rsp.RegistryId
				log.Printf("%s registered as %s", vars.Name, rsp.RegistryId)
			}
		}
	}

	if vars.ConfigServer != "" {
		go publishConfigs(node, vars)
	}
	return nil
}

func StopNode(node Node, vars *Vars) error {
	if vars.Registry != "" {
		regClient, err := registryClient(vars)
		if err != nil {
			return fmt.Errorf("could not connect to registry server: %s", err)
		}
		_, err = regClient.Deregister(context.Background(), &servicepb.DeregisterRequest{RegistryId: vars.RegistryID})
		return err
	}
	return nil
}

func publishConfigs(node Node, vars *Vars) {
	vars.configChanInput = make(chan conf.Map, 1)
	node.ShareState(vars.configChanInput)
	for {
		if vars.configClient == nil {
			for {
				conn, err := net.GRPCMutualTlsDial(vars.ConfigServer, vars.authorityCert, vars.serviceCert, vars.serviceKey)
				if err != nil {
					log.Printf("could not create secure connection to config server: %s\n", err)
					<-time.After(time.Second * 3)
					continue
				}
				vars.configClient = configpb.NewConfigClient(conn)
				break
			}
		}

		state, ok := <-vars.configChanInput
		if !ok {
			log.Println("stopped sharing state")
			return
		}

		stateBytes, err := json.Marshal(state)
		if err != nil {
			log.Fatalf("could not encode state: %s\n", err)
		}

		_, err = vars.configClient.Set(context.Background(), &configpb.SetRequest{
			Key:   "",
			Value: stateBytes,
		})
		if err != nil {
			log.Fatalf("could not set config: %s\n", err)
		}
	}
}
