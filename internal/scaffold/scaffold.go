package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type TemplateData struct {
	ProjectName    string
	Description    string
	Date           string
	Language       string
	Framework      string
	PackageManager string
	TestRunner     string
}

type Scaffolder struct {
	fs     fs.FS
	dir    string
	data   TemplateData
	force  bool
	dryRun bool
}

func New(embedFS fs.FS, dir string, data TemplateData, force, dryRun bool) *Scaffolder {
	data.Date = time.Now().Format("2006-01-02")
	return &Scaffolder{fs: embedFS, dir: dir, data: data, force: force, dryRun: dryRun}
}

// manifest define el orden de creación (importante para el output).
var manifest = []string{
	"CLAUDE.md",
	"AGENTS.md",
	"init.sh",
	"feature_list.json",
	"claude-progress.md",
	"session-handoff.md",
	"clean-state-checklist.md",
	"evaluator-rubric.md",
	"quality-document.md",
	"docs/ARCHITECTURE.md",
	"docs/PRODUCT.md",
	"docs/RELIABILITY.md",
	"data/design-notes.md",
	"data/retrieval-plan.md",
	"agents/planner.md",
	"agents/reviewer.md",
	"agents/worker.md",
}

func (s *Scaffolder) Run() (created, skipped []string, err error) {
	files := append([]string(nil), manifest...)

	// archOverride mapea destino -> fuente cuando el template específico existe
	archOverride := ""
	if s.data.Language != "" && s.data.Language != "generic" {
		archFile := fmt.Sprintf("docs/architecture_%s_%s.md", s.data.Language, s.data.Framework)
		if _, fsErr := fs.Stat(s.fs, archFile); fsErr == nil {
			archOverride = archFile
		}
	}

	for _, rel := range files {
		dest := filepath.Join(s.dir, rel)

		if _, statErr := os.Stat(dest); statErr == nil && !s.force {
			skipped = append(skipped, rel)
			continue
		}

		if s.dryRun {
			created = append(created, rel)
			continue
		}

		if mkErr := os.MkdirAll(filepath.Dir(dest), 0755); mkErr != nil {
			return nil, nil, mkErr
		}

		src := rel
		if rel == "docs/ARCHITECTURE.md" && archOverride != "" {
			src = archOverride
		}

		content, readErr := fs.ReadFile(s.fs, src)
		if readErr != nil {
			return nil, nil, readErr
		}

		var out []byte
		if bytes.Contains(content, []byte("{{")) {
			tmpl, tmplErr := template.New(rel).Parse(string(content))
			if tmplErr != nil {
				return nil, nil, tmplErr
			}
			var buf bytes.Buffer
			if execErr := tmpl.Execute(&buf, s.data); execErr != nil {
				return nil, nil, execErr
			}
			out = buf.Bytes()
		} else {
			out = content
		}

		if writeErr := os.WriteFile(dest, out, 0644); writeErr != nil {
			return nil, nil, writeErr
		}

		if strings.HasSuffix(rel, "init.sh") && runtime.GOOS != "windows" {
			_ = os.Chmod(dest, 0755)
		}

		created = append(created, rel)
	}

	if !s.dryRun && s.data.Language != "" {
		state := fmt.Sprintf("PROJECT_TYPE=%s\nFRAMEWORK=%s\nPACKAGE_MANAGER=%s\nTEST_RUNNER=%s\n",
			s.data.Language, s.data.Framework, s.data.PackageManager, s.data.TestRunner)
		_ = os.WriteFile(filepath.Join(s.dir, ".harness-state"), []byte(state), 0644)
	}

	return created, skipped, nil
}
