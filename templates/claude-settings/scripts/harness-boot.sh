#!/bin/bash
# Hook UserPromptSubmit: ejecuta init.sh una vez por sesión en entornos harness.
# Si init.sh falla, emite HARNESS_BOOT_FAILED: y sale con error para bloquear al agente.
LOCK="/tmp/harness-boot-$(echo "$PWD" | md5sum | cut -c1-8).done"
[ -f "$LOCK" ] && exit 0

# Salir silencioso si no es entorno harness
[ -f ".harness-state" ] || [ -f "feature_list.json" ] || exit 0

if [ ! -f "init.sh" ]; then
  echo "HARNESS_BOOT_FAILED: init.sh no encontrado en $(pwd). El entorno no está inicializado."
  exit 1
fi

if ! bash init.sh 2>&1; then
  echo "HARNESS_BOOT_FAILED: init.sh terminó con error en $(pwd). Resuelve el error antes de continuar."
  exit 1
fi

touch "$LOCK"
