package configs

import (
	"github.com/zoenion/common/prompt"
)

type Access struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
}

func (access *Access) Prompt() error {
	prompt.Header("Access")

	var err error

	access.Access, err = prompt.TextWithDefault("Access", access.Access, false)
	if err != nil {
		return err
	}

	access.Secret, err = prompt.Password("Secret")
	return err
}
