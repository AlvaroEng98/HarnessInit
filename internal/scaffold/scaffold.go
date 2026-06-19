package scaffold

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type TemplateData struct {
	ProjectName string
	Description string
	Date        string
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
	for _, rel := range manifest {
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

		content, readErr := fs.ReadFile(s.fs, rel)
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
	return created, skipped, nil
}
