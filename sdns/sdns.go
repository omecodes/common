package sdns

import (
	"context"
	"github.com/zoenion/common/errors"
	servicepb "github.com/zoenion/common/proto/service"
	"google.golang.org/grpc"
	"time"
)

func ResolveAddress(reg string, namespace string, name string, protocol string) (string, error) {
	rc, err := servicepb.NewClient(reg, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	rsp, err := rc.Get(ctx, &servicepb.GetRequest{})
	if err != nil {
		return "", err
	}

	for _, n := range rsp.Info.Nodes {
		if n.Protocol == protocol {
			return n.Address, nil
		}
	}

	return "", errors.NotFound
}
