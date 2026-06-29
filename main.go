package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed .claude docs AGENT.md CLAUDE.md feature_list.json init.sh session-handoff.md progress CHECKPOINTS.md .gitignore
var templateFS embed.FS

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "version", "--version", "-v":
		fmt.Printf("harness v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: harness <command>

Commands:
  init [--client claude|opencode] [directory]    Scaffold project (default: both clients)
  version                                         Print version
  help                                            Show this help`)
}

func cmdInit() {
	printBanner()

	detected := detectTools()
	chosen := selectClient(detected)
	if chosen == "" {
		os.Exit(1)
	}
	_ = chosen // TODO: usar en lógica de creación de estructura

	args := os.Args[2:]
	target := "."
	client := "both"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--client":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --client requires a value (claude|opencode)\n")
				os.Exit(1)
			}
			i++
			client = args[i]
			if client != "claude" && client != "opencode" {
				fmt.Fprintf(os.Stderr, "Error: --client must be claude or opencode\n")
				os.Exit(1)
			}
		default:
			target = args[i]
		}
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(absTarget)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(absTarget, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		for _, e := range entries {
			if e.Name() == "AGENT.md" || e.Name() == "feature_list.json" {
				fmt.Fprintf(os.Stderr, "Error: target already contains a harness project\n")
				os.Exit(1)
			}
		}
	}

	wantClaude := client == "claude" || client == "both"
	wantOpencode := client == "opencode" || client == "both"

	err = fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		switch {
		case path == ".claude":
			if wantClaude {
				return os.MkdirAll(filepath.Join(absTarget, path), 0755)
			}
			return nil // still walk into it to reach agent files

		case path == ".claude/agents":
			if wantClaude {
				return os.MkdirAll(filepath.Join(absTarget, path), 0755)
			}
			return nil // still walk into it to reach agent files

		case strings.HasPrefix(path, ".claude/agents/"):
			data, err := templateFS.ReadFile(path)
			if err != nil {
				return err
			}
			if wantClaude {
				dest := filepath.Join(absTarget, path)
				if err := os.WriteFile(dest, data, 0644); err != nil {
					return err
				}
				fmt.Printf("  Created %s\n", filepath.ToSlash(path))
			}
			if wantOpencode {
				dir := filepath.Join(absTarget, ".opencode", "agents")
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, d.Name()), data, 0644); err != nil {
					return err
				}
				fmt.Printf("  Created .opencode/agents/%s\n", d.Name())
			}
			return nil

		default:
			destPath := filepath.Join(absTarget, path)
			if d.IsDir() {
				return os.MkdirAll(destPath, 0755)
			}
			data, err := templateFS.ReadFile(path)
			if err != nil {
				return err
			}
			mode := fs.FileMode(0644)
			if d.Name() == "init.sh" {
				mode = 0755
			}
			if err := os.WriteFile(destPath, data, mode); err != nil {
				return err
			}
			fmt.Printf("  Created %s\n", filepath.ToSlash(path))
			return nil
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	emptyDirs := []string{"src", "tests", "specs"}
	for _, dir := range emptyDirs {
		dirPath := filepath.Join(absTarget, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not create %s/\n", filepath.ToSlash(dirPath))
		} else {
			fmt.Printf("  Created %s/\n", filepath.ToSlash(dirPath))
		}
	}

	fmt.Println()
	fmt.Printf("Done! Harness project scaffolded in %s\n", absTarget)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit feature_list.json with your project info")
	fmt.Println("  2. Run ./init.sh to verify the environment")
	fmt.Println("  3. Read AGENT.md to understand the workflow")
}
