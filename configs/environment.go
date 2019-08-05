package configs

import (
	"github.com/zoenion/common/conf"
)

type Environment struct {
	Databases Databases `json:"databases"`
	Network   *Network  `json:"network"`
}

func (environment *Environment) Prompt() error {
	var err error
	environment.Databases = map[string]conf.Map{}

	err = environment.Databases.Prompt()
	if err != nil {
		return err
	}

	environment.Network = new(Network)
	return environment.Network.Prompt()
}
