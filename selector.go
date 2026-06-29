package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	colorCyan  = "\033[1;36m"
	colorDim   = "\033[2m"
	colorReset = "\033[0m"
)

var knownTools = []struct {
	name   string
	binary string
}{
	{"Claude Code", "claude"},
	{"Cursor", "cursor"},
	{"OpenCode", "opencode"},
	{"Pi", "pi"},
}

type detectedTool struct {
	name   string
	binary string
	found  bool
}

func printBanner() {
	art := colorCyan +
		"    _    ____  ____  ___ _     \n" +
		"   / \\  |  _ \\|  _ \\|_ _| |   \n" +
		"  / _ \\ | |_) | |_) || || |   \n" +
		" / ___ \\|  __/|  _ < | || |___\n" +
		"/_/   \\_\\_|   |_| \\_\\___|_____|\n" +
		colorReset
	fmt.Println(art)
	fmt.Printf(colorDim+"  Project scaffolding for AI-assisted development  v%s"+colorReset+"\n\n", version)
}

func detectTools() []detectedTool {
	result := make([]detectedTool, len(knownTools))
	for i, t := range knownTools {
		_, err := exec.LookPath(t.binary)
		result[i] = detectedTool{t.name, t.binary, err == nil}
	}
	return result
}

// selectClient muestra las herramientas detectadas y retorna el binary elegido.
// Retorna "" si el usuario cancela o no hay herramientas disponibles.
func selectClient(detected []detectedTool) string {
	fmt.Println("  Herramientas AI detectadas:\n")
	for _, d := range detected {
		if d.found {
			fmt.Printf("    \033[32m✓\033[0m  %s\n", d.name)
		} else {
			fmt.Printf("    \033[90m✗  %s\033[0m\n", d.name)
		}
	}
	fmt.Println()

	var available []detectedTool
	for _, d := range detected {
		if d.found {
			available = append(available, d)
		}
	}

	if len(available) == 0 {
		fmt.Fprintln(os.Stderr, "  \033[31mNo se detectó ninguna herramienta compatible.\033[0m")
		fmt.Fprintln(os.Stderr, "  Instala claude, cursor, opencode o pi y vuelve a intentar.")
		return ""
	}

	if len(available) == 1 {
		fmt.Printf("  \033[1;36m→ Estructura para: %s (auto-detectado)\033[0m\n\n", available[0].name)
		return available[0].binary
	}

	fmt.Println("  ¿Para qué herramienta crear la estructura de carpetas?\n")
	for i, d := range available {
		fmt.Printf("  \033[1m[%d]\033[0m %s\n", i+1, d.name)
	}
	fmt.Printf("\n  Selecciona [1-%d] (Enter para cancelar): ", len(available))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("\n  Cancelado.")
		return ""
	}

	n, err := strconv.Atoi(input)
	if err != nil || n < 1 || n > len(available) {
		fmt.Fprintln(os.Stderr, "\n  \033[31mSelección inválida.\033[0m")
		return ""
	}

	chosen := available[n-1]
	fmt.Printf("\n  \033[1;36m→ Estructura para: %s\033[0m\n\n", chosen.name)
	return chosen.binary
}
