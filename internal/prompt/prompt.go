package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/term"
)

// Resolve devuelve nombre y descripción desde flags o interactivamente.
// En modo no-interactivo (sin TTY), falla si name está vacío.
func Resolve(name, description string) (string, string, error) {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	if name == "" && !isTTY {
		return "", "", fmt.Errorf("--name requerido en modo no-interactivo")
	}

	if name == "" {
		if err := survey.AskOne(&survey.Input{Message: "Nombre del proyecto:"}, &name, survey.WithValidator(survey.Required)); err != nil {
			return "", "", err
		}
	}

	if description == "" && isTTY {
		if err := survey.AskOne(&survey.Input{
			Message: "Descripción (Enter para usar el nombre):",
			Default: name,
		}, &description); err != nil {
			return "", "", err
		}
	}

	if description == "" {
		description = name
	}

	return name, description, nil
}
