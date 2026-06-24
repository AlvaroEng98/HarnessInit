package scaffold

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/alvaroeng98/HarnessInit/templates"
)

// TestManifestFilesAreEmbedded verifica que cada entrada del manifest existe en el
// embed FS; si no, init/upgrade fallaría en runtime al leer la plantilla.
func TestManifestFilesAreEmbedded(t *testing.T) {
	for _, rel := range manifest {
		if _, err := fs.ReadFile(templates.FS, rel); err != nil {
			t.Errorf("manifest incluye %q pero no está embebido: %v", rel, err)
		}
	}
}

// TestNoGhostTemplates verifica que todo fichero embebido bajo docs/ y data/ está en el
// manifest. Atrapa ghost files (embebidos pero nunca generados, como el antiguo
// architecture_python_fastapi.md).
func TestNoGhostTemplates(t *testing.T) {
	inManifest := make(map[string]bool, len(manifest))
	for _, rel := range manifest {
		inManifest[rel] = true
	}

	for _, root := range []string{"docs", "data"} {
		err := fs.WalkDir(templates.FS, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !inManifest[path] {
				t.Errorf("fichero embebido %q no está en el manifest (ghost file)", path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir(%q): %v", root, err)
		}
	}
}

// TestHarnessOwnedSubsetOfManifest verifica que toda clave de harnessOwned aparece en el
// manifest; una clave huérfana sería un no-op silencioso en Upgrade().
func TestHarnessOwnedSubsetOfManifest(t *testing.T) {
	inManifest := make(map[string]bool, len(manifest))
	for _, rel := range manifest {
		inManifest[rel] = true
	}
	for rel := range harnessOwned {
		if !inManifest[rel] {
			t.Errorf("harnessOwned incluye %q que no está en el manifest", rel)
		}
	}
}

// TestDestOverrideKeysInManifest verifica que las claves de destOverride existen en el
// manifest (un override de una ruta inexistente no se aplica nunca).
func TestDestOverrideKeysInManifest(t *testing.T) {
	inManifest := make(map[string]bool, len(manifest))
	for _, rel := range manifest {
		inManifest[rel] = true
	}
	for rel := range destOverride {
		if !inManifest[rel] {
			t.Errorf("destOverride mapea %q que no está en el manifest", rel)
		}
	}
}

// TestScriptsAreExecutableTargets documenta que los scripts del manifest terminan en .sh
// para que materialize() los marque ejecutables.
func TestShScriptsRecognized(t *testing.T) {
	for _, rel := range manifest {
		if strings.HasSuffix(rel, "init.sh") || strings.HasSuffix(rel, "harness-boot.sh") {
			return
		}
	}
	t.Error("ningún script .sh en el manifest; materialize() no marcaría ejecutables")
}
