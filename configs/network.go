package configs

import (
	"github.com/zoenion/common/prompt"
)

type Network struct {
	PrivateAddress string `json:"private_address"`
	PublicAddress  string `json:"public_address"`
	PublicDomain   string `json:"public_domain"`
}

func (n *Network) Prompt() error {

	prompt.Header("Network")

	var err error

	defaultPrivateAddress := n.PrivateAddress
	if defaultPrivateAddress == "" {
		defaultPrivateAddress = "127.0.0.1"
	}

	n.PrivateAddress, err = prompt.TextWithDefault("Private address", defaultPrivateAddress, false)
	if err != nil {
		return err
	}

	n.PublicAddress, err = prompt.TextWithDefault("Public address", n.PublicAddress, false)
	if err != nil {
		return err
	}

	n.PublicDomain, err = prompt.TextWithDefault("Public domain", n.PublicDomain, true)
	return err
}
