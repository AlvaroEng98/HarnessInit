package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/term"
)

// Resolve devuelve nombre, descripción, lenguaje y framework desde flags o interactivamente.
// En modo no-interactivo (sin TTY), falla si name o language están vacíos.
func Resolve(name, description, language, framework string) (string, string, string, string, error) {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	if (name == "" || language == "") && !isTTY {
		return "", "", "", "", fmt.Errorf("--name y --language requeridos en modo no-interactivo")
	}

	if name == "" {
		if err := survey.AskOne(&survey.Input{Message: "Nombre del proyecto:"}, &name, survey.WithValidator(survey.Required)); err != nil {
			return "", "", "", "", err
		}
	}

	if description == "" && isTTY {
		if err := survey.AskOne(&survey.Input{
			Message: "Descripción (Enter para usar el nombre):",
			Default: name,
		}, &description); err != nil {
			return "", "", "", "", err
		}
	}

	if description == "" {
		description = name
	}

	if language == "" {
		if err := survey.AskOne(&survey.Select{
			Message: "Lenguaje:",
			Options: []string{"python", "node", "go", "java-maven", "java-gradle", "generic"},
		}, &language); err != nil {
			return "", "", "", "", err
		}
	}

	if framework == "" && isTTY {
		switch language {
		case "python":
			if err := survey.AskOne(&survey.Select{
				Message: "Framework:",
				Options: []string{"fastapi", "django", "flask", "none"},
			}, &framework); err != nil {
				return "", "", "", "", err
			}
		case "node":
			if err := survey.AskOne(&survey.Select{
				Message: "Framework:",
				Options: []string{"nextjs", "nestjs", "express", "none"},
			}, &framework); err != nil {
				return "", "", "", "", err
			}
		default:
			framework = "none"
		}
	}

	if framework == "" {
		framework = "none"
	}

	return name, description, language, framework, nil
}
