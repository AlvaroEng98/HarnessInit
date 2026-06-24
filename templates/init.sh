#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

# Modo de ejecución:
#   bootstrap (default): instala dependencias + sanity ligero. Lo que corre el hook en cada sesión.
#   verify             : corre la suite de verificación completa. Bajo demanda (./init.sh verify).
MODE="${1:-bootstrap}"
if [[ "$MODE" != "bootstrap" && "$MODE" != "verify" ]]; then
  echo "ERROR: modo '$MODE' inválido. Usa: bootstrap | verify" >&2
  exit 1
fi

if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "WARNING: No se encontró repositorio git." >&2
  echo "         'git log' fallará en los pasos de sesión." >&2
  echo "         Ejecuta: git init && git add -A && git commit -m 'init'" >&2
fi

# ---------------------------------------------------------------------------
# Comandos del proyecto (configúralos manualmente)
# ---------------------------------------------------------------------------
# Sustituye cada REPLACE: por el comando real de tu stack.
# Ejemplos:
#   INSTALL_CMD=(npm install)        VERIFY_CMD=(npm test)        START_CMD=(npm run dev)
#   INSTALL_CMD=(uv sync)            VERIFY_CMD=(pytest)          START_CMD=(python main.py)
#   INSTALL_CMD=(go mod download)    VERIFY_CMD=(go test ./...)   START_CMD=(go run .)

# REPLACE: comando para instalar dependencias
INSTALL_CMD=(echo "REPLACE: configura INSTALL_CMD en init.sh")
# REPLACE: comando de verificación / tests
VERIFY_CMD=(echo "REPLACE: configura VERIFY_CMD en init.sh")
# REPLACE: comando para arrancar la app
START_CMD=(echo "REPLACE: configura START_CMD en init.sh")

# ---------------------------------------------------------------------------
# Ejecución
# ---------------------------------------------------------------------------

echo "==> Working directory: $PWD"

echo "==> Syncing dependencies"
"${INSTALL_CMD[@]}"

if [[ "$MODE" == "verify" ]]; then
  echo "==> Running baseline verification"
  "${VERIFY_CMD[@]}"
else
  echo "==> Bootstrap completo. Verificación bajo demanda: ./init.sh verify"
fi

echo "==> Startup command"
printf '    %q' "${START_CMD[@]}"
printf '\n'

if [ "${RUN_START_COMMAND:-0}" = "1" ]; then
  echo "==> Starting the app"
  exec "${START_CMD[@]}"
fi

echo "Set RUN_START_COMMAND=1 if you want init.sh to launch the app directly."
