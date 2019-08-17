package service

import (
	"context"
	"errors"
	servicepb "github.com/zoenion/common/proto/service"
	"google.golang.org/grpc"
)

func resolveService(v *Vars, serviceName string) (*servicepb.Info, error) {
	if v.Registry == "" {
		return nil, errors.New("no registry server address specified in vars")
	}

	if v.loaded.registryClient == nil {
		var err error
		v.loaded.registryClient, err = servicepb.NewClient(v.Registry, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
	}

	rsp, err := v.loaded.registryClient.Get(context.Background(), &servicepb.GetRequest{
		RegistryId: v.Namespace + "::" + serviceName,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Info, nil
}