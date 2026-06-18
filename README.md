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
go install github.com/alvaroeng98/harness-init@latest
```

## Uso

```sh
harness-init init --name "mi-proyecto" --description "Descripción del proyecto"
```

Todos los flags son opcionales. Sin `--name`, el CLI los solicita de forma interactiva.

| Flag | Por defecto | Descripción |
|------|-------------|-------------|
| `--name` | interactivo | Nombre del proyecto |
| `--description` | igual que `--name` | Descripción corta |
| `--dir` | `.` | Directorio destino |
| `--force` | `false` | Sobreescribir ficheros existentes |
| `--dry-run` | `false` | Mostrar ficheros sin crearlos |

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

Los ficheros que contienen `{{` se renderizan como plantillas Go con `ProjectName`, `Description` y `Date`.

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
