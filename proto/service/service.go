package servicepb

import (
	"google.golang.org/grpc"
)

func NewClient(serverAddr string, opts ...grpc.DialOption) (RegistryClient, error) {
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, err
	}
	return NewRegistryClient(conn), nil
}
