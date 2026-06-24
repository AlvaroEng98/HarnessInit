package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "harness-init",
	Short:   "Inicializa la estructura de harness LLM en un proyecto",
	Version: version,
}

func Execute() {
	updateCh := checkUpdateAsync()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	printUpdateNotice(updateCh)
}

func checkUpdateAsync() chan string {
	ch := make(chan string, 1)
	if version == "dev" || os.Getenv("HARNESS_NO_UPDATE_CHECK") == "1" || isUpdateCommand() {
		close(ch)
		return ch
	}
	go func() {
		latest, err := fetchLatestVersion()
		if err != nil || latest == "v"+version {
			close(ch)
			return
		}
		ch <- latest
	}()
	return ch
}

func printUpdateNotice(ch chan string) {
	select {
	case v, ok := <-ch:
		if ok {
			fmt.Fprintf(os.Stderr, "\nNueva versión disponible: %s → ejecuta: harness-init update\n", v)
		}
	case <-time.After(1500 * time.Millisecond):
	}
}

func isUpdateCommand() bool {
	return len(os.Args) > 1 && os.Args[1] == "update"
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(upgradeCmd)
}
