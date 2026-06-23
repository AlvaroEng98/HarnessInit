# CLAUDE.md — HarnessTeam Orchestrator

Estás trabajando en un repositorio diseñado para trabajo de implementación de larga duración.
Prioriza la completación fiable, la continuidad entre sesiones y la verificación explícita
sobre la velocidad.

Flujo: **Planner → (aprobación usuario) → Worker → Reviewer → [loop si REQUEST_CHANGES]**

## Bucle Operacional

Si ves un mensaje que contiene `HARNESS_BOOT_FAILED:` en tu contexto, detente
inmediatamente. Reporta el error exacto al usuario. No ejecutes ninguna tarea.

Al comienzo de cada sesión:

1. Ejecuta `./init.sh`. **Si falla (exit code ≠ 0), DETENTE INMEDIATAMENTE.** No leas ningún fichero.
   No ejecutes ninguna tarea. Reporta el error exacto al usuario y espera instrucciones.
2. Ejecuta `pwd` y confirma que estás en la raíz del repositorio esperada.
3. Lee `claude-progress.md`.
4. Lee `feature_list.json`.
5. Revisa los commits recientes con `git log --oneline -5`.
6. Lee `.harness-state` para obtener `PROJECT_TYPE`, `FRAMEWORK`, `PACKAGE_MANAGER` y `TEST_RUNNER`.
   - Si `PACKAGE_MANAGER=uv` → instala dependencias con `uv add <pkg>`. Nunca `uv pip install` ni `pip install`.
   - Si `TEST_RUNNER=pytest` → escribe y ejecuta tests con `pytest`. Nunca `python -m unittest`.
   Lee `docs/ARCHITECTURE.md`. Ver **Contrato de Arquitectura** más abajo.
7. Verifica si la ruta de smoke o end-to-end de referencia ya está rota.

Luego selecciona exactamente una característica inacabada y trabaja solo en esa característica hasta
que la verifiques o documentes por qué está bloqueada.

## Contrato de Arquitectura

`docs/ARCHITECTURE.md` es la fuente de verdad para estructura de directorios, capas y sus restricciones.

**Reglas operativas:**

1. Antes de crear un fichero nuevo → consulta la tabla de capas en `ARCHITECTURE.md` para determinar dónde va.
2. Antes de añadir lógica a una capa → verifica su columna "Prohibido" en `ARCHITECTURE.md`.
3. No crees directorios fuera de la estructura definida en `ARCHITECTURE.md`.
4. Si el caso no está contemplado en `ARCHITECTURE.md` → consulta al usuario antes de improvisar.

**Jerarquía de documentos:**

| Decisión | Fuente autoritativa |
|----------|---------------------|
| Estructura de directorios y asignación de capas | `ARCHITECTURE.md` |
| Flujo del agente (Planner / Worker / Reviewer) | `CLAUDE.md` |
| Contradicción entre ambos | Notificar al usuario antes de actuar |

## Reglas

- Una característica activa a la vez.
- No afirmes completación sin evidencia ejecutable.
- No reescribas la lista de características para ocultar trabajo inacabado.
- No elimines o debilites tests solo para hacer que la tarea parezca completa.
- Usa los artefactos del repositorio como el sistema de registro.

## Reglas de Delegación

| Condición | Acción |
|-----------|--------|
| Tarea trivial, 1 archivo | Implementa inline sin agentes |
| 2+ archivos no triviales | Lanza Planner → espera `planner-plan.v1` → lanza Worker |
| Después de cualquier Worker | Siempre lanza Reviewer en contexto fresco |
| Reviewer devuelve `APPROVED` | Actualizar `feature_list.json`, cerrar sesión |
| Reviewer devuelve `REQUEST_CHANGES` | Re-lanzar Worker con findings del Reviewer como contexto |
| Reviewer devuelve `BLOCKED` | Escalar al usuario. No reintentar automáticamente |

**Regla crítica**: El orchestrator nunca implementa código directamente si la tarea toca 2+ archivos. Delegar siempre.

## Contratos de Resultado

### planner-plan.v1

El Planner devuelve este JSON antes de que el Worker pueda comenzar.

```json
{
  "schema_version": "planner-plan.v1",
  "task_id": "...",
  "summary": "descripción de la tarea en una línea",
  "plan_table": [
    {"file": "ruta/al/archivo", "purpose": "qué hace", "depends_on": [], "priority": 1}
  ],
  "validation_commands": ["npm test", "bash tests/smoke.sh"],
  "risks": ["descripción de riesgos identificados"]
}
```

### worker-report.v1

El Worker devuelve este JSON al terminar su ciclo de implementación.

```json
{
  "schema_version": "worker-report.v1",
  "task_id": "...",
  "files_modified": ["ruta/archivo1", "ruta/archivo2"],
  "tests_run": ["npm test", "bash tests/smoke.sh"],
  "test_result": "pass | fail | skip",
  "evidence": ["línea de output relevante 1", "línea de output relevante 2"]
}
```

### review-result.v1

El Reviewer devuelve este JSON. Es el único artefacto que determina el cierre de una tarea.

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVED | REQUEST_CHANGES | BLOCKED",
  "findings": [
    {
      "severity": "critical | major | minor",
      "file": "ruta/archivo",
      "line": null,
      "description": "descripción concreta del hallazgo"
    }
  ],
  "rationale": "justificación del veredicto"
}
```

## Archivos Requeridos

- `feature_list.json`
- `claude-progress.md`
- `init.sh`
- `session-handoff.md` cuando una entrega compacta es útil

## Puerta de Completación

Una característica puede pasar a `passing` solo después de que la verificación requerida tenga éxito
y el resultado esté registrado.

## Antes de Detenerte

1. Actualiza `claude-progress.md` con el estado verificado más reciente.
2. Actualiza `feature_list.json` con el nuevo estado de la tarea.
3. Registra riesgos o bloqueos sin resolver.
4. Deja el repo en estado reiniciable desde `./init.sh`.
