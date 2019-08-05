package configs

import (
	"fmt"
	"github.com/zoenion/common/prompt"
)

type OnionApp struct {
	AuthorityCertPath string `json:"authority_cert_path"`
	Access            Access `json:"access"`
	Registry          string `json:"registry"`
}

func (oc *OnionApp) Prompt() (err error) {
	fmt.Println()
	fmt.Println()
	fmt.Println("===================")
	fmt.Println("Onion configuration")
	fmt.Println()

	fmt.Println()
	fmt.Println("Service discovery registry")
	fmt.Println("--------------------------")
	oc.Registry, err = prompt.TextWithDefault("Address", "127.0.0.1:9777", false)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Authority Certificate")
	fmt.Println("---------------------")
	oc.AuthorityCertPath, err = prompt.TextWithDefault("Path", oc.AuthorityCertPath, false)
	if err != nil {
		return
	}

	err = oc.Access.Prompt()
	return
}
