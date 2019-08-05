package configs

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/prompt"
)

type Databases map[string]conf.Map

func (dbs Databases) Prompt() error {

	prompt.Header("Databases")
	var err error

	for {
		choice, err := prompt.Selection("Set database config?", []string{"MySQL", "Mongo", "Redis", "No"})
		if err != nil {
			return err
		}

		if choice == "No" {
			break
		}

		name, err := prompt.Text("Config name", false)
		if err != nil {
			return err
		}

		oldDbCfg := dbs[name]

		var dbCfg conf.Map
		if choice == "MySQL" {
			dbCfg, err = MysqlPrompt(oldDbCfg)

		} else if choice == "Mongo" {
			dbCfg, err = MongoPrompt(oldDbCfg)

		} else {
			dbCfg, err = RedisPrompt(oldDbCfg)
		}

		if err != nil {
			return err
		}

		dbs[name] = dbCfg
	}
	return err
}

func MysqlPrompt(old conf.Map) (conf.Map, error) {
	var (
		oldHost, oldUser, oldName, oldCharset, oldWrapper string
	)

	if old != nil {
		oldHost, _ = old.GetString("host")
		oldUser, _ = old.GetString("user")
		oldName, _ = old.GetString("name")
		oldCharset, _ = old.GetString("charset")
		oldWrapper, _ = old.GetString("wrapper")
	} else {
		oldHost = "localhost:3306"
		oldCharset = "utf8"
	}

	host, err := prompt.TextWithDefault("Host", oldHost, false)
	if err != nil {
		return nil, err
	}

	user, err := prompt.TextWithDefault("User", oldUser, false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	name, err := prompt.TextWithDefault("Name", oldName, false)
	if err != nil {
		return nil, err
	}

	encoding, err := prompt.TextWithDefault("Charset", oldCharset, false)
	if err != nil {
		return nil, err
	}

	wrapper, _ := prompt.TextWithDefault("Wrapper", oldWrapper, true)

	cfg := conf.Map{
		"type":     "sql",
		"driver":   "mysql",
		"host":     host,
		"user":     user,
		"password": password,
		"name":     name,
		"charset":  encoding,
	}
	hasWrapper := len(wrapper) > 0
	cfg["wrapped"] = hasWrapper
	if hasWrapper {
		cfg["wrapper"] = wrapper
	}
	return cfg, nil
}

func MongoPrompt(old conf.Map) (conf.Map, error) {
	host, err := prompt.TextWithDefault("Host", "localhost:27017", false)
	if err != nil {
		return nil, err
	}
	user, err := prompt.Text("User", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	name, err := prompt.Text("Name", false)
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"user":     user,
		"password": password,
		"name":     name,
	}, nil
}

func RedisPrompt(old conf.Map) (conf.Map, error) {
	host, err := prompt.TextWithDefault("Host", "localhost:6379", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"password": password,
	}, nil
}
