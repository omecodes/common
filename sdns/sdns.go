package sdns

import (
	"context"
	"github.com/zoenion/common/errors"
	registrypb "github.com/zoenion/common/proto/registry"
	"google.golang.org/grpc"
	"time"
)

func ResolveAddress(reg string, namespace string, name string, protocol string) (string, error) {
	rc, err := registrypb.NewClient(reg, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	rsp, err := rc.Get(ctx, &registrypb.GetRequest{})
	if err != nil {
		return "", err
	}

	for _, n := range rsp.Application.Nodes {
		if n.Protocol == protocol {
			return n.Address, nil
		}
	}

	return "", errors.NotFound
}
