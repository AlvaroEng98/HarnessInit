package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alvaroeng98/HarnessInit/internal/scaffold"
	"github.com/alvaroeng98/HarnessInit/templates"
	"github.com/spf13/cobra"
)

var (
	flagUpgradeDir    string
	flagUpgradeDryRun bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Actualiza los ficheros del harness en un proyecto existente, preservando tu estado",
	RunE:  runUpgrade,
}

func init() {
	upgradeCmd.Flags().StringVar(&flagUpgradeDir, "dir", ".", "directorio del proyecto a actualizar")
	upgradeCmd.Flags().BoolVar(&flagUpgradeDryRun, "dry-run", false, "mostrar cambios sin aplicarlos")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	// Detectar que es un proyecto harness: .harness-version es el marcador principal;
	// feature_list.json es el fallback para proyectos creados antes del versionado.
	if _, err := os.Stat(filepath.Join(flagUpgradeDir, scaffold.VersionFile)); err != nil {
		if _, err := os.Stat(filepath.Join(flagUpgradeDir, "feature_list.json")); err != nil {
			return fmt.Errorf("no es un proyecto harness (falta .harness-version y feature_list.json); usa 'harness-init init'")
		}
	}

	// En builds publicados, si ya está en la versión actual no hay nada que hacer.
	// En builds 'dev' siempre se re-aplica (útil para probar en local).
	existing := scaffold.ReadVersion(flagUpgradeDir)
	if version != "dev" && existing == version {
		fmt.Printf("El harness ya está en la versión %s.\n", version)
		return nil
	}

	abs, err := filepath.Abs(flagUpgradeDir)
	if err != nil {
		return err
	}
	name := filepath.Base(abs)

	s := scaffold.New(templates.FS, flagUpgradeDir, scaffold.TemplateData{
		ProjectName: name,
	}, false, flagUpgradeDryRun)

	updated, preserved, err := s.Upgrade()
	if err != nil {
		return err
	}

	for _, f := range updated {
		if flagUpgradeDryRun {
			fmt.Printf("[DRY-RUN] %s\n", f)
		} else {
			fmt.Printf("[UPDATE] %s\n", f)
		}
	}
	for _, f := range preserved {
		fmt.Printf("[KEEP]   %s\n", f)
	}

	if flagUpgradeDryRun {
		fmt.Printf("\n%d se actualizarían, %d preservados.\n", len(updated), len(preserved))
		return nil
	}

	if err := scaffold.WriteVersion(flagUpgradeDir, version); err != nil {
		return err
	}

	fmt.Printf("\n%d actualizados, %d preservados.\n", len(updated), len(preserved))
	fmt.Println("Revisa los cambios con: git diff")
	return nil
}
