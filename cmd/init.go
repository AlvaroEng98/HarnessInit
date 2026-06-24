package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/alvaroeng98/HarnessInit/internal/scaffold"
	"github.com/alvaroeng98/HarnessInit/templates"
	"github.com/spf13/cobra"
)

var (
	flagName   string
	flagDir    string
	flagForce  bool
	flagDryRun bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Crea los ficheros del harness en el directorio destino",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&flagName, "name", "", "nombre del proyecto (default = nombre del directorio destino)")
	initCmd.Flags().StringVar(&flagDir, "dir", ".", "directorio destino")
	initCmd.Flags().BoolVar(&flagForce, "force", false, "sobreescribir ficheros existentes")
	initCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "mostrar ficheros sin crearlos")
}

func runInit(cmd *cobra.Command, args []string) error {
	name := flagName
	if name == "" {
		abs, err := filepath.Abs(flagDir)
		if err != nil {
			return err
		}
		name = filepath.Base(abs)
	}

	s := scaffold.New(templates.FS, flagDir, scaffold.TemplateData{
		ProjectName: name,
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

	if err := scaffold.WriteVersion(flagDir, version); err != nil {
		return err
	}

	fmt.Printf("\n%d ficheros procesados (%d creados, %d omitidos).\n", total, len(created), len(skipped))
	if len(created) > 0 {
		fmt.Println("Ejecuta: ./init.sh")
	}

	return nil
}
