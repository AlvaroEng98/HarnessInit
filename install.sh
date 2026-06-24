#!/usr/bin/env sh
set -eu

REPO="AlvaroEng98/HarnessInit"
BINARY="harness-init"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Arquitectura no soportada: $ARCH" >&2; exit 1 ;;
esac

VERSION="${VERSION:-$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')}"

ARCHIVE="${BINARY}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

mkdir -p "$INSTALL_DIR"

echo "Instalando ${BINARY} ${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "/tmp/${ARCHIVE}"
tar -xzf "/tmp/${ARCHIVE}" -C /tmp "${BINARY}"
chmod +x "/tmp/${BINARY}"
mv "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Instalado en ${INSTALL_DIR}/${BINARY}"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "AVISO: ${INSTALL_DIR} no está en \$PATH. Añade esta línea a tu ~/.bashrc o ~/.zshrc:"
     echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
     ;;
esac

echo "Ejecuta: harness-init --help"
