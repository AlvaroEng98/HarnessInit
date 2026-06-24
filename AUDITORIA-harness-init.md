# Auditoría — harness-init

> Revisión de entorno harness para agentes IA · 24/06/2026
> Alcance: implementación SDD, delegación multi-agente, flujo de trabajo, ficheros generados, optimización de `CLAUDE.md`/`AGENTS.md`.

---

## Conclusión ejecutiva

**El harness está bien diseñado en el papel pero roto en su ruta feliz.** Un `harness-init init` recién ejecutado produce un proyecto que **no arranca**: el agente queda bloqueado en el primer prompt de la sesión. Tres defectos críticos lo causan, y los tres son corregibles sin rediseñar nada.

| # | Severidad | Defecto | Efecto |
|---|-----------|---------|--------|
| C1 | 🔴 Crítico | `init.go` hardcodea `python/fastapi/uv/pytest` | Todo proyecto se marca como Python → `init.sh` exige `.venv` → falla en Go/Node/Java |
| C2 | 🔴 Crítico | El hook ejecuta la suite completa (`pytest`/`go test`) en cada prompt | Proyecto nuevo sin tests → `init.sh` falla → `HARNESS_BOOT_FAILED` → agente bloqueado |
| C3 | 🔴 Crítico | Los agentes se generan en `agents/`, no en `.claude/agents/` | Claude Code no los descubre → frontmatter y `disallowedTools` **nunca se aplican** en runtime |
| M1 | 🟠 Mayor | 7 ficheros se generan vacíos y ninguna herramienta los rellena | Ruido permanente; el "fuente de verdad" `ARCHITECTURE.md` llega vacío o con stack equivocado |
| M2 | 🟠 Mayor | `CLAUDE.md` y `AGENTS.md` duplican la secuencia de arranque | Riesgo de drift; sesgo Python contradice el auto-detect de `init.sh` |
| m1 | 🟡 Menor | Plumbing muerto (`Language`/`Framework`) y fichero huérfano `templates/.harness-state` | Deuda del refactor que quitó los flags |

**Lo que sí funciona bien:** los contratos JSON (`planner-plan.v1`, `worker-report.v1`, `review-result.v1`), el patrón Return Envelope, la separación de roles Planner/Worker/Reviewer/Orchestrator, y la lógica de auto-detección de stack dentro de `init.sh` (que el resto del código sabotea).

---

## 1. Implementación del SDD

El harness implementa un SDD ligero y coherente **como proceso**:

```
feature_list.json   → inventario/spec de features (fuente de verdad de estado)
planner-plan.v1     → diseño (tabla de archivos + comandos de validación)
worker-report.v1    → implementación + evidencia ejecutable
review-result.v1    → puerta de aceptación (único artefacto que cierra tarea)
```

Esto es sólido: hay puerta de completación basada en evidencia, estados explícitos (`not_started/in_progress/blocked/passing`), y prohibición de marcar `passing` sin verificación. **Bien.**

**Brecha de fondo: el sustrato del spec llega vacío.** Un SDD necesita que la especificación exista antes del diseño. Aquí los documentos que deberían *contener* esa spec se generan en blanco:

- `docs/ARCHITECTURE.md` → **fuente de verdad declarada** en `CLAUDE.md` (§Contrato de Arquitectura), pero el template genérico tiene **0 bytes**.
- `docs/PRODUCT.md` → **0 bytes**.
- `docs/RELIABILITY.md` → **0 bytes**.

El SPEC interno afirma que un comando `configure` rellena estos ficheros ("Fase 2 ya implementada"). **Es falso:** los únicos comandos registrados son `init` y `update` (`cmd/root.go:67-68`). No existe `configure`. Por tanto los documentos de arquitectura **quedan vacíos para siempre**, y `CLAUDE.md` instruye al agente a tratar como "fuente de verdad" un fichero sin contenido.

> **Recomendación:** o se implementa el comando que puebla esos docs, o `init.sh` falla rápido cuando `ARCHITECTURE.md` está vacío (en vez de dejar al agente operar sobre una fuente de verdad inexistente), o se elimina la dependencia dura de `CLAUDE.md` sobre `ARCHITECTURE.md`.

---

## 2. Delegación multi-agente

Arquitectura de 4 agentes con frontmatter YAML, contratos y `disallowedTools`. La tabla de delegación del Orchestrator es clara y los roles están bien acotados (Planner read-only, Reviewer read-only + git de solo lectura, Worker RWE sin spawning). El Return Envelope evita que el contrato JSON se pierda si el último paso es una tool call. **Diseño correcto.**

**Defecto crítico C3 — los guardrails no se aplican.** Claude Code descubre sub-agentes en `.claude/agents/`. El scaffolder los escribe en `agents/` (raíz del proyecto):

- `internal/scaffold/scaffold.go:63-66` — `destOverride` solo remapea `claude-settings/*` a `.claude/`. Las entradas `agents/*.md` del manifest no tienen override → aterrizan en `<proyecto>/agents/`.
- Resultado: el runtime **no carga** estos ficheros como agentes. El frontmatter `disallowedTools: [Write, Edit, ...]` es texto inerte. Toda la propuesta de "guardrails declarativos a nivel runtime" del SPEC (Brechas 2 y 3) queda anulada por una ruta equivocada.
- Síntoma colateral: el propio SPEC dice `Ver .claude/agents/orchestrator.md`, pero el `CLAUDE.md` generado dice `Ver agents/orchestrator.md`. Inconsistencia que delata el bug.

> **Recomendación:** añadir a `destOverride` el mapeo `agents/*.md → .claude/agents/*.md` y actualizar las referencias en `CLAUDE.md`. Sin esto, la delegación depende 100% de que el LLM obedezca prosa, que es justo lo que el SPEC quería blindar.

---

## 3. Flujo de trabajo (arranque de sesión)

El flujo previsto: hook `UserPromptSubmit` → `harness-boot.sh` → `init.sh` (una vez por sesión vía lockfile). La intención —fail-fast si el entorno está roto— es correcta. La ejecución bloquea al usuario en el caso más común.

**C1 — hardcode de stack.** `cmd/init.go:41-44`:

```go
language := "python"
framework := "fastapi"
packageManager := "uv"
testRunner := "pytest"
```

El commit `0c429ec` ("remove language and framework flags") quitó los flags pero dejó estos valores fijos. Consecuencias en cadena:

1. `scaffold.go:137-140` escribe `.harness-state` con `PROJECT_TYPE=python` **siempre**.
2. `init.sh:18-19` preserva `PROJECT_TYPE` (semántica de "override manual") → cualquier proyecto se trata como Python.
3. `init.sh:213-219` exige `.venv` → en un proyecto Go/Node/Java no existe → `exit 1`.
4. `archOverride` (`scaffold.go:72-78`) usa `python/fastapi` → **todo proyecto recibe un `ARCHITECTURE.md` de FastAPI**, sea cual sea su stack.

La ironía: `init.sh` tiene auto-detección completa y correcta (Go/Node/Python/Java/Maven/Gradle, gestores y frameworks). El hardcode de `init.go` la cortocircuita.

**C2 — el hook corre la suite completa en cada prompt.** `harness-boot.sh:15` ejecuta `bash init.sh`, que corre `INSTALL_CMD` + `VERIFY_CMD` (`init.sh:235-238`). En un proyecto recién scaffoldeado sin tests:

- `pytest` sin tests devuelve exit code 5 → con `set -euo pipefail`, `init.sh` falla.
- `harness-boot.sh` convierte ese fallo en `HARNESS_BOOT_FAILED:`.
- `CLAUDE.md` §Bucle Operacional ordena: "detente inmediatamente, no ejecutes ninguna tarea".

Es decir: **el proyecto que el propio harness genera no puede arrancar su primera sesión.** Además, ejecutar la suite completa en *cada* `UserPromptSubmit` (no solo la primera) sería lento y frágil incluso si pasara.

> **Recomendaciones flujo:**
> - Separar *bootstrap* (instalar deps + sanity check ligero) de *verificación* (suite completa, bajo demanda). El hook no debe correr todos los tests en cada prompt.
> - Tratar "no se recogieron tests" (pytest exit 5, `go test` sin ficheros) como no-fatal durante el arranque.
> - Eliminar el hardcode: dejar que `init.sh` detecte, y que `init.go` no escriba `PROJECT_TYPE` salvo que el usuario lo fuerce explícitamente.

---

## 4. Ficheros generados — innecesarios / inesperados

`harness-init init` produce 20 ficheros. Varios son ruido:

| Fichero | Estado | Problema |
|---------|--------|----------|
| `docs/PRODUCT.md` | 0 bytes | Vacío, sin herramienta que lo rellene |
| `docs/RELIABILITY.md` | 0 bytes | Íd. |
| `docs/ARCHITECTURE.md` (genérico) | 0 bytes | Íd., y declarado "fuente de verdad" |
| `quality-document.md` | ~0 | Vacío |
| `session-handoff.md` | ~0 | Vacío |
| `data/design-notes.md` | ~0 | Vacío; concepto RAG irrelevante para la mayoría de proyectos |
| `data/retrieval-plan.md` | ~0 | Íd. |

Siete ficheros vacíos en cada proyecto = ruido que ni el agente ni el usuario saben rellenar. El directorio `data/` (notas de diseño + plan de recuperación) sugiere un caso de uso RAG que no aplica a un harness genérico.

**Ficheros del propio repo (no del scaffold):**

- `templates/.harness-state` — **huérfano**: no está en `embed.go` (`templates/embed.go:5`) ni en el `manifest`. El scaffolder genera su propio `.harness-state` al final. Este fichero no se usa nunca.
- `elecctions.md` (raíz) — notas sueltas (árbol de decisión Python). Gitignored, pero ensucia el working tree. Es scratch.
- `SPEC-harnessinit-agent-upgrades.md` — gitignored; útil como histórico pero no debería vivir en raíz.

**Plumbing muerto (m1):** tras quitar los flags, `scaffold.go` mantiene los campos `Language/Framework/PackageManager/TestRunner` en `TemplateData` y la lógica `archOverride`, alimentados solo por el hardcode. Es deuda del refactor.

> **Recomendación:** generar los docs vacíos solo si hay contenido real que poner (o con un placeholder de una línea que diga qué escribir), no como ficheros 0-byte. Borrar `templates/.harness-state` y `elecctions.md`. Mover el SPEC a `docs/` o a un directorio de diseño.

---

## 5. Optimización de `CLAUDE.md` y `AGENTS.md`

**Problema raíz: duplicación.** Ambos ficheros repiten casi la misma secuencia de arranque (ejecutar `init.sh`, `pwd`, leer `claude-progress.md`, leer `feature_list.json`, `git log -5`). Dos fuentes para la misma instrucción → drift garantizado.

Reparto correcto de responsabilidades:
- `CLAUDE.md` lo lee Claude Code automáticamente. Debe ser la **única** fuente del protocolo operativo.
- `AGENTS.md` es el estándar para *otros* agentes/herramientas. Debe ser un puntero fino a `CLAUDE.md`, no una copia.

**Sesgo Python en `CLAUDE.md` §Bucle Operacional paso 6:**

```
Si PACKAGE_MANAGER=uv → uv add ...
Si TEST_RUNNER=pytest → pytest ...
```

Contradice el diseño multi-stack de `init.sh`. Debe leer `.harness-state` de forma genérica (usar `INSTALL_CMD`/`VERIFY_CMD` resueltos por `init.sh`) en vez de codificar herramientas Python.

**`AGENTS.md` paso 6** lee `smoke_test` de `feature_list.json`, que en un scaffold fresco es `"REPLACE: ..."` → el agente debe detenerse y pedir configuración. Correcto, pero combinado con C1/C2 significa que el arranque siempre se detiene en proyecto nuevo. Conviene un mensaje único y claro de "primer uso: configura `smoke_test` y stack".

> **Recomendaciones concretas:**
> 1. Consolidar la secuencia de arranque en `CLAUDE.md`; reducir `AGENTS.md` a un puntero ("El protocolo operativo vive en `CLAUDE.md`. Reglas de trabajo abajo.").
> 2. Quitar referencias a `uv`/`pytest` de `CLAUDE.md`; delegar comandos a lo que resuelva `init.sh`.
> 3. Corregir la ruta de agentes a `.claude/agents/` en `CLAUDE.md`.
> 4. Añadir una sección "Primer uso" que liste los placeholders `REPLACE:` que el usuario debe rellenar antes de la primera sesión productiva.

---

## 6. Plan de corrección priorizado

```
1. C1  Quitar hardcode python en init.go      → verify: init en repo Go no escribe PROJECT_TYPE=python
2. C3  destOverride agents/ → .claude/agents/ → verify: ls .claude/agents/ tras init muestra los 4 .md
3. C2  Bootstrap ligero en hook; tests aparte → verify: init.sh en proyecto sin tests termina exit 0
4. M1  No generar docs 0-byte; borrar huérfanos→ verify: init no crea ficheros vacíos
5. M2  Deduplicar CLAUDE.md/AGENTS.md          → verify: una sola secuencia de arranque
6. m1  Limpiar plumbing Language/Framework     → verify: go test ./... pasa
```

Los tres críticos (C1/C2/C3) son los que convierten al harness de "no usable out-of-the-box" a "funcional". El resto es pulido.

---

## Anexo — Notas de seguridad (fuera de alcance principal)

- `install.sh` / `install.ps1` usan `curl | sh` sin verificación de checksum (`grep -c sha256 install.sh` → 0). Riesgo de cadena de suministro si el repo o la CDN se comprometen. Recomendado: pinning de versión + verificación SHA-256.
- `cmd/update.go:downloadAndExtract` descarga y ejecuta un binario de releases de GitHub **sin verificar checksum ni firma**. Mismo riesgo en el canal de auto-update.
