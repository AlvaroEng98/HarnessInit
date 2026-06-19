#!/usr/bin/env sh
set -eu

BINARY="harness-init"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TARGET="${INSTALL_DIR}/${BINARY}"

if [ ! -f "$TARGET" ]; then
  echo "${BINARY} no encontrado en ${INSTALL_DIR}"
  exit 0
fi

if [ -w "$INSTALL_DIR" ]; then
  rm "$TARGET"
else
  sudo rm "$TARGET"
fi

echo "Desinstalado: ${TARGET}"
