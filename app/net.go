package app

import "fmt"

func (v *Vars) GRPCAddress() string {
	addr := v.Domain
	if addr == "" {
		addr = v.IP
	}
	return fmt.Sprintf("%s:%s", addr, v.GatewayGRPCPort)
}

func (v *Vars) HTTPAddress() string {
	addr := v.Domain
	if addr == "" {
		addr = v.IP
	}
	return fmt.Sprintf("%s:%s", addr, v.GatewayHTTPPort)
}
