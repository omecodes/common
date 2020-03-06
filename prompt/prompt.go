package prompt

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/manifoldco/promptui"
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

func Selection(label string, values ...string) (string, error) {
	return selection(label, values)
}
