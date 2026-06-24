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
	"agents/orchestrator.md",
	"agents/planner.md",
	"agents/reviewer.md",
	"agents/worker.md",
	"claude-settings/settings.json",
	"claude-settings/scripts/harness-boot.sh",
}

// destOverride mapea src del embed -> destino real (para renombrar directorios).
var destOverride = map[string]string{
	"claude-settings/settings.json":           ".claude/settings.json",
	"claude-settings/scripts/harness-boot.sh": ".claude/scripts/harness-boot.sh",
	"agents/orchestrator.md":                  ".claude/agents/orchestrator.md",
	"agents/planner.md":                       ".claude/agents/planner.md",
	"agents/reviewer.md":                      ".claude/agents/reviewer.md",
	"agents/worker.md":                        ".claude/agents/worker.md",
}

// VersionFile marca, en el proyecto destino, la versión del harness con la que se generó.
const VersionFile = ".harness-version"

// harnessOwned son los ficheros de protocolo/tooling que pertenecen al harness y que
// Upgrade() sobrescribe. El resto del manifest contiene estado del usuario y se preserva.
var harnessOwned = map[string]bool{
	"CLAUDE.md":                               true,
	"AGENTS.md":                               true,
	"clean-state-checklist.md":                true,
	"evaluator-rubric.md":                     true,
	"quality-document.md":                     true,
	"agents/orchestrator.md":                  true,
	"agents/planner.md":                       true,
	"agents/reviewer.md":                      true,
	"agents/worker.md":                        true,
	"claude-settings/settings.json":           true,
	"claude-settings/scripts/harness-boot.sh": true,
}

// destFor devuelve la ruta destino (relativa) aplicando destOverride.
func destFor(rel string) string {
	if override, ok := destOverride[rel]; ok {
		return override
	}
	return rel
}

func (s *Scaffolder) Run() (created, skipped []string, err error) {
	for _, rel := range manifest {
		destRel := destFor(rel)
		dest := filepath.Join(s.dir, destRel)

		if _, statErr := os.Stat(dest); statErr == nil && !s.force {
			skipped = append(skipped, destRel)
			continue
		}

		if s.dryRun {
			created = append(created, destRel)
			continue
		}

		if mkErr := s.materialize(rel, dest); mkErr != nil {
			return nil, nil, mkErr
		}

		created = append(created, destRel)
	}

	return created, skipped, nil
}

// Upgrade sobrescribe solo los ficheros del harness (harnessOwned) con las plantillas
// actuales y preserva el resto (estado del usuario). No depende de force ni de existencia.
func (s *Scaffolder) Upgrade() (updated, preserved []string, err error) {
	for _, rel := range manifest {
		destRel := destFor(rel)
		if !harnessOwned[rel] {
			preserved = append(preserved, destRel)
			continue
		}

		if s.dryRun {
			updated = append(updated, destRel)
			continue
		}

		if mkErr := s.materialize(rel, filepath.Join(s.dir, destRel)); mkErr != nil {
			return nil, nil, mkErr
		}

		updated = append(updated, destRel)
	}

	return updated, preserved, nil
}

// materialize lee el fichero embebido rel, interpola si tiene tags de plantilla y lo escribe
// en dest (creando directorios y marcando como ejecutable los scripts .sh).
func (s *Scaffolder) materialize(rel, dest string) error {
	if mkErr := os.MkdirAll(filepath.Dir(dest), 0755); mkErr != nil {
		return mkErr
	}

	content, readErr := fs.ReadFile(s.fs, rel)
	if readErr != nil {
		return readErr
	}

	var out []byte
	if bytes.Contains(content, []byte("{{")) {
		tmpl, tmplErr := template.New(rel).Parse(string(content))
		if tmplErr != nil {
			return tmplErr
		}
		var buf bytes.Buffer
		if execErr := tmpl.Execute(&buf, s.data); execErr != nil {
			return execErr
		}
		out = buf.Bytes()
	} else {
		out = content
	}

	if writeErr := os.WriteFile(dest, out, 0644); writeErr != nil {
		return writeErr
	}

	if runtime.GOOS != "windows" && (strings.HasSuffix(rel, "init.sh") || strings.HasSuffix(rel, "harness-boot.sh")) {
		_ = os.Chmod(dest, 0755)
	}

	return nil
}

// WriteVersion escribe el marcador de versión del harness en el proyecto destino.
func WriteVersion(dir, v string) error {
	return os.WriteFile(filepath.Join(dir, VersionFile), []byte(v+"\n"), 0644)
}

// ReadVersion lee el marcador de versión; devuelve "" si no existe o no se puede leer.
func ReadVersion(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, VersionFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
