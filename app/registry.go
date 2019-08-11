package app

import (
	registrypb "github.com/zoenion/common/proto/registry"
	"google.golang.org/grpc"
)

func registryClient(v *Vars) (registrypb.RegistryClient, error) {
	if v.registryClient != nil {
		return v.registryClient, nil
	}

	var err error

	v.registryClient, err = registrypb.NewClient(v.Registry, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return v.registryClient, nil
}
