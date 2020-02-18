package prompt

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

func unixNumber(label string, defaultValue string, masked bool) (int64, error) {
	number, e := unixText(fmt.Sprintf("%s ", label), defaultValue, false, masked)
	if e != nil {
		return -1, nil
	}
	return strconv.ParseInt(number, 10, 64)
}

func unixText(label string, defaultValue string, acceptEmpty bool, masked bool) (string, error) {
	validate := func(txt string) error {
		if len(txt) == 0 && !acceptEmpty {
			return errors.BadInput
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s ", label),
		Validate:  validate,
		AllowEdit: true,
		Default:   defaultValue,
	}
	if masked {
		prompt.Mask = '*'
	}
	return prompt.Run()
}

func unixSelect(label string, values []string) (string, error) {
	prompt := promptui.Select{
		Label: fmt.Sprintf("%s ", label),
		Items: values,
	}
	_, result, err := prompt.Run()
	return result, err
}

func winNumber(label string, defaultValue string, masked bool) (int64, error) {
	number, e := winText(fmt.Sprintf("%s", label), defaultValue, false, masked)
	if e != nil {
		return 0, e
	}
	return strconv.ParseInt(number, 10, 64)
}

func winText(label string, defaultValue string, acceptEmpty bool, masked bool) (string, error) {
	var text string
	var questions []*survey.Question
	if masked {
		questions = []*survey.Question{
			{
				Validate: func(text interface{}) error {
					str, _ := text.(string)
					if len(str) == 0 && !acceptEmpty {
						return errors.BadInput
					}
					return nil
				},
				Name:   "text",
				Prompt: &survey.Password{Message: fmt.Sprintf("%s:", label)},
			},
		}
	} else {
		questions = []*survey.Question{
			{
				Name:   "text",
				Prompt: &survey.Input{Message: fmt.Sprintf("%s:", label), Default: defaultValue},
			},
		}
	}
	err := survey.Ask(questions, &text)
	return text, err
}

func winSelect(label string, values []string) (string, error) {
	var result string
	dbSelect := &survey.Select{
		Message: fmt.Sprintf("%s:", label),
		Options: values,
	}
	err := survey.AskOne(dbSelect, &result, nil)
	return result, err
}

func number(label string, defaultValue string, masked bool) (int64, error) {
	if runtime.GOOS == "windows" {
		return winNumber(label, defaultValue, masked)
	}
	return unixNumber(label, defaultValue, masked)
}

func text(label string, defaultValue string, canBeEmpty bool, masked bool) (string, error) {
	if runtime.GOOS == "windows" {
		return winText(label, defaultValue, canBeEmpty, masked)
	}
	return unixText(label, defaultValue, canBeEmpty, masked)
}

func selection(label string, values []string) (string, error) {
	if runtime.GOOS == "windows" {
		return winSelect(label, values)
	}
	return unixSelect(label, values)
}

func Text(label string, canBeEmpty bool) (string, error) {
	return text(label, "", canBeEmpty, false)
}

func Integer(label string) (int64, error) {
	return number(label, "", false)
}

func IntegerWithDefaultValue(label string, defaultValue int64) (int64, error) {
	return number(label, fmt.Sprintf("%d", defaultValue), false)
}

func TextWithDefault(label, defaultValue string, canBeEmpty bool) (string, error) {
	return text(label, defaultValue, canBeEmpty, false)
}

func Password(label string) (string, error) {
	return text(label, "", false, true)
}

func PasswordNumber(label string) (int64, error) {
	return number(label, "", true)
}

func Selection(label string, values []string) (string, error) {
	return selection(label, values)
}

func Access(title string, defaultValue string) (conf.Map, error) {
	if title != "" {
		Header(title)
	}

	cfg := conf.Map{}
	access, err := text("Key", defaultValue, false, false)
	if err != nil {
		return nil, err
	}

	secret, err := text("Secret", "", false, true)
	if err != nil {
		return nil, err
	}

	cfg["access"] = access
	cfg["secret"] = secret
	return cfg, nil
}

func ServiceIP(label, defaultValue string) (string, error) {
	if label == "" {
		label = "Set the address to bind services to"
	}
	return text(label, defaultValue, false, false)
}

func PublicIP(label, defaultValue string) (string, error) {
	if label == "" {
		label = "Set the external IP(for public certificates)"
	}
	return text(label, defaultValue, false, false)
}

func NetworkDomain(label, defaultValue string) (string, error) {
	if label == "" {
		label = "Set the app domain name"
	}
	return text(label, defaultValue, false, false)
}

func Registry(title string, defaultValue string) (conf.Map, error) {
	if title != "" {
		Header(title)
	}

	host, err := text("Address", defaultValue, false, false)
	if err != nil {
		return nil, err
	}
	return conf.Map{
		"host": host,
	}, nil
}

func Mysql() (conf.Map, error) {
	host, err := text("host", "localhost:3306", false, false)
	if err != nil {
		return nil, err
	}
	user, err := text("user", "", false, false)
	if err != nil {
		return nil, err
	}

	password, err := text("password", "", false, true)
	if err != nil {
		return nil, err
	}

	name, err := text("name", "", false, false)
	if err != nil {
		return nil, err
	}

	encoding, err := text("charset", "utf8", false, false)
	if err != nil {
		return nil, err
	}

	wrapper, _ := text("wrapper", "", true, false)

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

func Mongo() (conf.Map, error) {
	host, err := text("host", "localhost:27017", false, false)
	if err != nil {
		return nil, err
	}
	user, err := text("user", "", false, false)
	if err != nil {
		return nil, err
	}

	password, err := text("password", "", false, true)
	if err != nil {
		return nil, err
	}

	name, err := text("name", "", false, false)
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

func Redis() (conf.Map, error) {
	host, err := text("host", "localhost:6379", false, false)
	if err != nil {
		return nil, err
	}

	password, err := text("password", "", false, true)
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"password": password,
	}, nil
}

func Mailer() (conf.Map, error) {
	server, err := text("SMTP server", "", false, false)
	if err != nil {
		return nil, err
	}

	port, err := number("port", "", false)
	if err != nil {
		return nil, err
	}

	user, err := text("user", "", false, false)
	if err != nil {
		return nil, err
	}

	password, err := text("password", "", false, true)
	if err != nil {
		return nil, err
	}

	cfg := conf.Map{
		"server":   server,
		"port":     port,
		"user":     user,
		"password": password,
	}
	return cfg, nil
}

func DatabaseConfigs() (conf.Map, error) {
	cfg := conf.Map{}

	for {
		choice, err := selection("Set database config?", []string{"MySQL", "Mongo", "Redis", "No"})
		if err != nil {
			return nil, err
		}

		if choice == "No" {
			break
		}

		var dbCfg conf.Map
		if choice == "MySQL" {
			dbCfg, err = Mysql()
		} else if choice == "Mongo" {
			dbCfg, err = Mongo()
		} else {
			dbCfg, err = Redis()
		}

		if err != nil {
			return nil, err
		}

		label, err := text("Label for this config", "", false, false)
		if err != nil {
			return nil, err
		}
		cfg[label] = dbCfg
	}
	return cfg, nil
}

func Header(title string) {
	fmt.Println()
	fmt.Println()
	fmt.Println(title)
	fmt.Println()
}
