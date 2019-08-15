package configs

import (
	"fmt"
	"github.com/zoenion/common/prompt"
)

type App struct {
	Name                     string       `json:"service"`
	Environment              *Environment `json:"environment"`
	Mailer                   *Mailer      `json:"mailer"`
	Registry                 string       `json:"registry"`
	AuthorityCertificatePath string       `json:"authority_certificate_path"`
	Access                   *Access      `json:"access"`
}

func (app *App) Prompt(skipAuth bool) error {
	fmt.Println("***********************************")
	fmt.Println()
	fmt.Println("WELCOME TO THE CONFIGURATION WIZARD")
	fmt.Println()
	fmt.Println("***********************************")

	fmt.Println()
	fmt.Println()
	fmt.Println("================")
	fmt.Println("Application info")
	fmt.Println()

	var err error
	app.Name, err = prompt.Text("Name", false)
	if err != nil {
		return err
	}

	app.Environment = new(Environment)
	err = app.Environment.Prompt()
	if err != nil {
		return err
	}

	app.Mailer = new(Mailer)
	err = app.Mailer.Prompt()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Service discovery registry")
	fmt.Println("--------------------------")
	app.Registry, err = prompt.TextWithDefault("Address", "127.0.0.1:9777", false)
	if err != nil {
		return err
	}

	if !skipAuth {
		fmt.Println()
		fmt.Println("Authority Certificate")
		fmt.Println("---------------------")
		app.AuthorityCertificatePath, err = prompt.Text("Path", false)
		if err != nil {
			return err
		}

		app.Access = new(Access)
		return app.Access.Prompt()
	}
	return nil
}
