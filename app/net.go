package app

import "fmt"

func (a *Vars) GRPCAddress() string {
	addr := a.Domain
	if addr == "" {
		addr = a.IP
	}
	return fmt.Sprintf("%s:%s", addr, a.GatewayGRPCPort)
}

func (a *Vars) HTTPAddress() string {
	addr := a.Domain
	if addr == "" {
		addr = a.IP
	}
	return fmt.Sprintf("%s:%s", addr, a.GatewayHTTPPort)
}
