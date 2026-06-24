package scaffold_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alvaroeng98/HarnessInit/internal/scaffold"
	"github.com/alvaroeng98/HarnessInit/templates"
)

func TestRun_CreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{
		ProjectName: "TestProj",
	}, false, false)

	created, skipped, err := s.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("esperado 0 omitidos, got %d", len(skipped))
	}
	if len(created) != 20 {
		t.Errorf("esperado 20 creados, got %d: %v", len(created), created)
	}
}

func TestRun_SkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	// segunda pasada sin --force
	s2 := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	created, skipped, err := s2.Run()
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 0 {
		t.Errorf("esperado 0 creados en segunda pasada, got %d", len(created))
	}
	if len(skipped) != 20 {
		t.Errorf("esperado 20 omitidos, got %d", len(skipped))
	}
}

func TestRun_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "Orig"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	s2 := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "Nuevo"}, true, false)
	created, skipped, err := s2.Run()
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 20 {
		t.Errorf("esperado 20 con --force, got %d", len(created))
	}
	if len(skipped) != 0 {
		t.Errorf("esperado 0 omitidos con --force, got %d", len(skipped))
	}
}

func TestRun_InterpolatesFeatureList(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{
		ProjectName: "MiProyecto",
	}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "feature_list.json"))
	if err != nil {
		t.Fatal(err)
	}

	var fl map[string]any
	if err := json.Unmarshal(data, &fl); err != nil {
		t.Fatalf("feature_list.json no es JSON válido: %v", err)
	}

	if fl["project"] != "MiProyecto" {
		t.Errorf("campo project = %q, want MiProyecto", fl["project"])
	}
}

func TestRun_InitShExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permisos no aplican en Windows")
	}
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, "init.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("init.sh no es ejecutable: %v", info.Mode())
	}
}

func TestRun_HarnessBootExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permisos no aplican en Windows")
	}
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, ".claude", "scripts", "harness-boot.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("harness-boot.sh no es ejecutable: %v", info.Mode())
	}
}

func TestRun_DryRunCreatesNothing(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, true)
	created, _, err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 20 {
		t.Errorf("esperado 20 en dry-run, got %d", len(created))
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("dry-run creó %d ficheros, esperado 0", len(entries))
	}
}

func TestRun_GeneratesAgentsInClaudeDir(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"orchestrator.md", "planner.md", "reviewer.md", "worker.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", name)); err != nil {
			t.Errorf(".claude/agents/%s no existe: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "agents")); !os.IsNotExist(err) {
		t.Errorf("el directorio agents/ no debería existir en la raíz del proyecto")
	}
}

func TestUpgrade_OverwritesOwnedPreservesState(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	// El usuario modifica un fichero de estado y uno del harness.
	featurePath := filepath.Join(dir, "feature_list.json")
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(featurePath, []byte(`{"project":"P","mi_progreso":"NO BORRAR"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, []byte("CONTENIDO MODIFICADO"), 0644); err != nil {
		t.Fatal(err)
	}

	s2 := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	updated, preserved, err := s2.Upgrade()
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 11 {
		t.Errorf("esperado 11 actualizados, got %d: %v", len(updated), updated)
	}
	if len(preserved) != 9 {
		t.Errorf("esperado 9 preservados, got %d: %v", len(preserved), preserved)
	}

	// feature_list.json se preserva.
	feature, _ := os.ReadFile(featurePath)
	if !bytes.Contains(feature, []byte("NO BORRAR")) {
		t.Errorf("feature_list.json fue sobrescrito; debía preservarse")
	}

	// CLAUDE.md se restaura desde la plantilla.
	claude, _ := os.ReadFile(claudePath)
	if bytes.Equal(claude, []byte("CONTENIDO MODIFICADO")) {
		t.Errorf("CLAUDE.md no fue actualizado; debía sobrescribirse")
	}
}

func TestWriteReadVersion(t *testing.T) {
	dir := t.TempDir()

	if v := scaffold.ReadVersion(dir); v != "" {
		t.Errorf("dir sin marcador debe devolver \"\", got %q", v)
	}

	if err := scaffold.WriteVersion(dir, "v1.2.3"); err != nil {
		t.Fatal(err)
	}
	if v := scaffold.ReadVersion(dir); v != "v1.2.3" {
		t.Errorf("ReadVersion = %q, want v1.2.3", v)
	}
}

func TestRun_DoesNotWriteHarnessState(t *testing.T) {
	dir := t.TempDir()
	s := scaffold.New(templates.FS, dir, scaffold.TemplateData{ProjectName: "P"}, false, false)
	if _, _, err := s.Run(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".harness-state")); !os.IsNotExist(err) {
		t.Errorf("el scaffold no debe escribir .harness-state")
	}
}
