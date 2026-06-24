# harness-init

CLI en Go que inicializa la estructura de harness LLM en cualquier proyecto.

Un solo comando crea los ficheros de contexto, plantillas de sesión y documentación estructurada que necesita un agente de IA para trabajar de forma ordenada en un repositorio.

## Instalación

**Linux / macOS**

```sh
curl -fsSL https://raw.githubusercontent.com/alvaroeng98/harness-init/main/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/alvaroeng98/harness-init/main/install.ps1 | iex
```

**Desde el código fuente**

```sh
go install github.com/alvaroeng98/HarnessInit@latest
```

## Desinstalación

**Linux / macOS**

```sh
curl -fsSL https://raw.githubusercontent.com/alvaroeng98/harness-init/main/uninstall.sh | sh
```

> Instala en `~/.local/bin`, no requiere `sudo`. Si ese directorio no está en `$PATH`, el script lo indica al finalizar.

## Uso

```sh
harness-init init --name "mi-proyecto"
```

Todos los flags son opcionales. El CLI no hace preguntas: sin `--name`, usa el nombre del
directorio destino.

| Flag | Por defecto | Descripción |
|------|-------------|-------------|
| `--name` | nombre del directorio destino | Nombre del proyecto |
| `--dir` | `.` | Directorio destino |
| `--force` | `false` | Sobreescribir ficheros existentes |
| `--dry-run` | `false` | Mostrar ficheros sin crearlos |

### Actualizar un proyecto ya inicializado

Cuando publicas una versión nueva del harness, los proyectos creados con una versión anterior
pueden actualizarse sin perder tu trabajo:

```sh
harness-init upgrade --dir "mi-proyecto"
```

`upgrade` **sobrescribe solo los ficheros del harness** (protocolo y tooling: `CLAUDE.md`,
`AGENTS.md`, `.claude/agents/*`, `.claude/settings.json`, `.claude/scripts/harness-boot.sh`,
rúbricas y checklists) y **preserva tu estado** (`feature_list.json`, `init.sh` con tus comandos,
`claude-progress.md`, `session-handoff.md`, `docs/`, `data/`). Escribe la versión aplicada en
`.harness-version`; si el proyecto ya está en la versión actual, no hace nada.

| Flag | Por defecto | Descripción |
|------|-------------|-------------|
| `--dir` | `.` | Directorio del proyecto a actualizar |
| `--dry-run` | `false` | Mostrar cambios sin aplicarlos |

Revisa el resultado con `git diff` antes de confirmar.

## Ficheros generados

```
CLAUDE.md                    # Instrucciones para el agente LLM
AGENTS.md                    # Reglas de comportamiento multi-agente
init.sh                      # Script de inicialización de entorno
feature_list.json            # Inventario de funcionalidades
claude-progress.md           # Seguimiento de progreso de sesión
session-handoff.md           # Traspaso entre sesiones
clean-state-checklist.md     # Checklist de estado limpio
evaluator-rubric.md          # Rúbrica de evaluación
quality-document.md          # Documento de calidad
docs/ARCHITECTURE.md         # Arquitectura técnica
docs/PRODUCT.md              # Visión de producto
docs/RELIABILITY.md          # Fiabilidad y operaciones
data/design-notes.md         # Notas de diseño
data/retrieval-plan.md       # Plan de recuperación de datos
```

Los ficheros que contienen `{{` se renderizan como plantillas Go con `ProjectName` y `Date`.

## Desarrollo

```sh
# Compilar
make build

# Tests
make test

# Compilar para todas las plataformas
make build-all
```

Releases automáticos via [GoReleaser](https://goreleaser.com) + GitHub Actions al hacer push de un tag `v*`.

## Licencia

MIT
