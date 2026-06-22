# AGENTS.md

Este repositorio está diseñado para trabajo de agentes de codificación de larga duración. El objetivo no
es maximizar la salida bruta de código. El objetivo es dejar el repositorio en un estado donde la
próxima sesión pueda continuar sin adivinar.

## Flujo de Trabajo de Inicio

Antes de escribir código:

1. Ejecuta `./init.sh`. **Si falla (exit code ≠ 0), DETENTE INMEDIATAMENTE.** Reporta el error exacto y espera instrucciones.
2. Confirma el directorio de trabajo con `pwd`.
3. Lee `claude-progress.md` para el estado verificado más reciente y el próximo paso.
4. Lee `feature_list.json` y elige la característica inacabada de mayor prioridad.
5. Revisa los commits recientes con `git log --oneline -5`.
6. Lee `feature_list.json` → campo `smoke_test`. Ejecuta ese comando exacto. Si el campo dice `REPLACE:`, detente y pide al usuario que lo configure antes de continuar.

Si la verificación de referencia ya está fallando, corrígela primero. No apiles trabajo de
características nuevas sobre un estado inicial roto.

## Reglas de Trabajo

- Trabaja en una característica a la vez.
- No marques una característica como completa solo porque se añadió código.
- Mantén los cambios dentro del alcance de la característica seleccionada a menos que un bloqueo
  fuerce una corrección de soporte estrecha.
- No cambies silenciosamente las reglas de verificación durante la implementación.
- Prefiere artefactos duraderos del repositorio sobre resúmenes de chat.

## Artefactos Requeridos

- `feature_list.json`: fuente de verdad para el estado de las características
- `claude-progress.md`: registro de sesión y estado verificado actual
- `init.sh`: ruta estándar de inicio y verificación
- `session-handoff.md`: entrega compacta opcional para sesiones más grandes

## Definición de Completado

Una característica está hecha solo cuando todo lo siguiente es cierto:

- el comportamiento objetivo está implementado
- la verificación requerida realmente se ejecutó
- la evidencia está registrada en `feature_list.json` o `claude-progress.md`
- el repositorio permanece reiniciable desde la ruta de inicio estándar

## Fin de Sesión

Antes de terminar una sesión:

1. Actualiza `claude-progress.md`.
2. Actualiza `feature_list.json`.
3. Registra cualquier riesgo o bloqueo sin resolver.
4. Deja el repositorio lo suficientemente limpio para que la próxima sesión pueda ejecutar
   `./init.sh` inmediatamente.