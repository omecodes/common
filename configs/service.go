package configs

import (
	"fmt"
	"github.com/zoenion/common/prompt"
)

type Service struct {
	App      App    `json:"app"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Database string `json:"database"`
}

func (service *Service) Prompt(app *App) error {
	var err error

	service.ID, err = prompt.TextWithDefault("identifier", service.ID, true)
	if err != nil {
		return err
	}

	service.Name, err = prompt.TextWithDefault("Name", service.Name, false)
	if err != nil {
		return err
	}

	choice, err := prompt.Selection(fmt.Sprintf("Does '%s' need a database?", service.Name), []string{"Yes", "No"})
	if err != nil {
		return err
	}
	if choice == "Yes" {
		if app.Environment.Databases == nil || len(app.Environment.Databases) == 0 {
			return fmt.Errorf("%s need a database config", service.Name)
		}

		var dbNames []string
		for n, _ := range app.Environment.Databases {
			dbNames = append(dbNames, n)
		}
		service.Database, err = prompt.Selection("Select a database config", dbNames)
	}

	return nil
}
