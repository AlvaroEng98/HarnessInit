package cmd

import (
	"fmt"
	"os"

	"github.com/alvaroeng98/HarnessInit/internal/prompt"
	"github.com/alvaroeng98/HarnessInit/internal/scaffold"
	"github.com/alvaroeng98/HarnessInit/templates"
	"github.com/spf13/cobra"
)

var (
	flagName        string
	flagDescription string
	flagLanguage    string
	flagFramework   string
	flagDir         string
	flagForce       bool
	flagDryRun      bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Crea los ficheros del harness en el directorio destino",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&flagName, "name", "", "nombre del proyecto")
	initCmd.Flags().StringVar(&flagDescription, "description", "", "descripción del proyecto (default = nombre)")
	initCmd.Flags().StringVar(&flagLanguage, "language", "", "lenguaje del proyecto")
	initCmd.Flags().StringVar(&flagFramework, "framework", "", "framework del proyecto")
	initCmd.Flags().StringVar(&flagDir, "dir", ".", "directorio destino")
	initCmd.Flags().BoolVar(&flagForce, "force", false, "sobreescribir ficheros existentes")
	initCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "mostrar ficheros sin crearlos")
}

func runInit(cmd *cobra.Command, args []string) error {
	name, description, language, framework, err := prompt.Resolve(flagName, flagDescription, flagLanguage, flagFramework)
	if err != nil {
		return err
	}

	packageManager := ""
	testRunner := ""
	if language == "python" {
		packageManager = "uv"
		testRunner = "pytest"
	}

	s := scaffold.New(templates.FS, flagDir, scaffold.TemplateData{
		ProjectName:    name,
		Description:    description,
		Language:       language,
		Framework:      framework,
		PackageManager: packageManager,
		TestRunner:     testRunner,
	}, flagForce, flagDryRun)

	created, skipped, err := s.Run()
	if err != nil {
		return err
	}

	for _, f := range created {
		if flagDryRun {
			fmt.Printf("[DRY-RUN] %s\n", f)
		} else {
			fmt.Printf("[CREATE] %s\n", f)
		}
	}
	for _, f := range skipped {
		fmt.Printf("[SKIP]   %s\n", f)
	}

	total := len(created) + len(skipped)
	if flagDryRun {
		fmt.Printf("\n%d ficheros se crearían (%d omitidos por existir).\n", len(created), len(skipped))
		return nil
	}

	fmt.Printf("\n%d ficheros procesados (%d creados, %d omitidos).\n", total, len(created), len(skipped))
	if len(created) > 0 {
		fmt.Println("Ejecuta: ./init.sh")
	}

	_ = os.Stderr
	return nil
}
