package app

import (
	"context"
	"errors"
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

func resolveService(v *Vars, serviceName string) (*registrypb.Application, error) {
	if v.Registry == "" {
		return nil, errors.New("no registry server address specified in vars")
	}

	if v.registryClient == nil {
		var err error
		v.registryClient, err = registrypb.NewClient(v.Registry, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
	}

	rsp, err := v.registryClient.Get(context.Background(), &registrypb.GetRequest{
		Id: v.Namespace + "::" + serviceName,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Application, nil
}
