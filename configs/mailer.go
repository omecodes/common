package configs

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/prompt"
)

type Mailer struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (mailer *Mailer) Prompt() error {

	prompt.Header("Mailer")
	var err error

	mailer.Host, err = prompt.TextWithDefault("Host", mailer.Host, false)
	if err != nil {
		return err
	}

	port, err := prompt.IntegerWithDefaultValue("Port", int64(mailer.Port))
	if err != nil {
		return nil
	}
	mailer.Port = int(port)

	mailer.User, err = prompt.TextWithDefault("User", mailer.User, false)
	if err != nil {
		return err
	}

	mailer.Password, err = prompt.Password("Password")
	if err != nil {
		return err
	}
	return err
}

func (mailer *Mailer) ToConf() conf.Map {
	cfg := conf.Map{}
	cfg["host"] = mailer.Host
	cfg["port"] = mailer.Port
	cfg["user"] = mailer.User
	cfg["password"] = mailer.Password
	return cfg
}
