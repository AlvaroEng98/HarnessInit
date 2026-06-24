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

# Cargar estado previo generado por harness-init (si existe)
# PROJECT_TYPE se preserva (puede ser override manual); el resto se re-detecta siempre.
if [[ -f .harness-state ]]; then
  # shellcheck source=/dev/null
  source .harness-state
  unset FRAMEWORK PACKAGE_MANAGER TEST_RUNNER INSTALL_CMD VERIFY_CMD START_CMD
fi

_VALID_PROJECT_TYPES="go node python java-maven java-gradle generic"
if [[ -n "${PROJECT_TYPE:-}" ]]; then
  if ! echo "$_VALID_PROJECT_TYPES" | grep -qw "$PROJECT_TYPE"; then
    echo "ERROR: PROJECT_TYPE='$PROJECT_TYPE' en .harness-state no es válido." >&2
    echo "       Valores aceptados: $_VALID_PROJECT_TYPES" >&2
    exit 1
  fi
fi

# ---------------------------------------------------------------------------
# Detección de tipo de proyecto
# ---------------------------------------------------------------------------

_detect_project_type() {
  if [[ -f go.mod ]]; then
    echo "go"
  elif [[ -f package.json ]]; then
    echo "node"
  elif [[ -f pom.xml ]]; then
    echo "java-maven"
  elif [[ -f build.gradle ]] || [[ -f build.gradle.kts ]]; then
    echo "java-gradle"
  elif [[ -f requirements.txt ]] || [[ -f pyproject.toml ]] || [[ -f setup.py ]]; then
    echo "python"
  else
    echo "generic"
  fi
}

_detect_python_framework() {
  if [[ -n "${FRAMEWORK:-}" ]]; then echo "$FRAMEWORK"; return; fi
  if [[ -f manage.py ]]; then
    echo "django"
  elif grep -qiE "^fastapi" requirements.txt 2>/dev/null || \
       grep -qi "fastapi" pyproject.toml 2>/dev/null; then
    echo "fastapi"
  elif grep -qiE "^flask" requirements.txt 2>/dev/null || \
       grep -qi "flask" pyproject.toml 2>/dev/null; then
    echo "flask"
  else
    echo "none"
  fi
}

_detect_python_dep_manager() {
  if [[ -n "${PACKAGE_MANAGER:-}" ]]; then echo "$PACKAGE_MANAGER"; return; fi
  if [[ -f uv.lock ]] || grep -q '\[tool\.uv\]' pyproject.toml 2>/dev/null; then
    echo "uv"
  elif [[ -f poetry.lock ]] || grep -q '\[tool\.poetry\]' pyproject.toml 2>/dev/null; then
    echo "poetry"
  elif [[ -f Pipfile ]]; then
    echo "pipenv"
  else
    echo "pip"
  fi
}

_detect_python_test_engine() {
  if [[ -n "${TEST_RUNNER:-}" ]]; then echo "$TEST_RUNNER"; return; fi
  if [[ -f pytest.ini ]] || [[ -f conftest.py ]] || \
     grep -q '\[tool\.pytest' pyproject.toml 2>/dev/null; then
    echo "pytest"
  elif grep -qiE "^pytest" requirements.txt 2>/dev/null; then
    echo "pytest"
  else
    echo "pytest"
  fi
}

_detect_node_package_manager() {
  if [[ -f yarn.lock ]]; then
    echo "yarn"
  elif [[ -f pnpm-lock.yaml ]]; then
    echo "pnpm"
  else
    echo "npm"
  fi
}

_detect_node_framework() {
  if [[ -n "${FRAMEWORK:-}" ]]; then echo "$FRAMEWORK"; return; fi
  if grep -q '"next"' package.json 2>/dev/null; then
    echo "nextjs"
  elif grep -q '"@nestjs/core"' package.json 2>/dev/null; then
    echo "nestjs"
  elif grep -q '"express"' package.json 2>/dev/null; then
    echo "express"
  else
    echo "none"
  fi
}

# ---------------------------------------------------------------------------
# Construcción de comandos según árbol de decisión
# ---------------------------------------------------------------------------

PROJECT_TYPE="${PROJECT_TYPE:-$(_detect_project_type)}"

case "$PROJECT_TYPE" in
  go)
    INSTALL_CMD=(go mod download)
    VERIFY_CMD=(go test ./...)
    START_CMD=(go run .)
    ;;

  node)
    _PM="$(_detect_node_package_manager)"
    _FW="$(_detect_node_framework)"

    case "$_PM" in
      yarn)  INSTALL_CMD=(yarn install) ;;
      pnpm)  INSTALL_CMD=(pnpm install) ;;
      *)     INSTALL_CMD=(npm install)  ;;
    esac

    case "$_PM" in
      yarn)  VERIFY_CMD=(yarn test)     ;;
      pnpm)  VERIFY_CMD=(pnpm test)     ;;
      *)     VERIFY_CMD=(npm test)      ;;
    esac

    case "$_FW" in
      nextjs) START_CMD=(npm run dev)         ;;
      nestjs) START_CMD=(npm run start:dev)   ;;
      *)
        case "$_PM" in
          yarn) START_CMD=(yarn start)  ;;
          pnpm) START_CMD=(pnpm start)  ;;
          *)    START_CMD=(npm start)   ;;
        esac
        ;;
    esac
    ;;

  python)
    _DEP="$(_detect_python_dep_manager)"
    _FW="$(_detect_python_framework)"
    _TEST="$(_detect_python_test_engine)"

    case "$_DEP" in
      uv)     INSTALL_CMD=(uv sync)                            ;;
      poetry) INSTALL_CMD=(poetry install)                     ;;
      pipenv) INSTALL_CMD=(pipenv install)                     ;;
      *)      INSTALL_CMD=(pip install -r requirements.txt)    ;;
    esac

    case "$_TEST" in
      pytest) VERIFY_CMD=(pytest) ;;
      *)      VERIFY_CMD=(python -m unittest discover) ;;
    esac

    case "$_FW" in
      fastapi) START_CMD=(uvicorn app.main:app --reload)  ;;
      flask)   START_CMD=(flask run)                      ;;
      django)  START_CMD=(python manage.py runserver)     ;;
      *)       START_CMD=(python main.py)                 ;;
    esac
    ;;

  java-maven)
    INSTALL_CMD=(mvn install -q -DskipTests)
    VERIFY_CMD=(mvn test)
    START_CMD=(mvn spring-boot:run)
    ;;

  java-gradle)
    INSTALL_CMD=(./gradlew build -q)
    VERIFY_CMD=(./gradlew test)
    START_CMD=(./gradlew bootRun)
    ;;

  *)
    INSTALL_CMD=(echo "TODO: configura INSTALL_CMD para este proyecto")
    VERIFY_CMD=(echo "TODO: configura VERIFY_CMD para este proyecto")
    START_CMD=(echo "TODO: configura START_CMD para este proyecto")
    ;;
esac

FRAMEWORK="${_FW:-${FRAMEWORK:-none}}"

# Persiste tipo/stack + comandos resueltos. CLAUDE.md consume INSTALL_CMD/VERIFY_CMD/START_CMD
# de forma genérica, sin asumir gestor concreto. %q los deja seguros para re-sourcing.
printf 'PROJECT_TYPE=%s\nFRAMEWORK=%s\nPACKAGE_MANAGER=%s\nTEST_RUNNER=%s\nINSTALL_CMD=%q\nVERIFY_CMD=%q\nSTART_CMD=%q\n' \
  "$PROJECT_TYPE" "$FRAMEWORK" "${_DEP:-}" "${_TEST:-}" \
  "${INSTALL_CMD[*]}" "${VERIFY_CMD[*]}" "${START_CMD[*]}" > .harness-state

# ---------------------------------------------------------------------------
# Ejecución
# ---------------------------------------------------------------------------

echo "==> Tipo detectado: $PROJECT_TYPE"
echo "==> Framework detectado: $FRAMEWORK"
echo "==> Working directory: $PWD"

if [[ "$PROJECT_TYPE" == "python" ]]; then
  if [[ ! -d .venv ]]; then
    echo "ERROR: No se encontró el directorio .venv." >&2
    echo "       Crea el entorno virtual antes de ejecutar init.sh:" >&2
    echo "         uv venv  |  python -m venv .venv  |  poetry config virtualenvs.in-project true && poetry install" >&2
    exit 1
  fi
  if [[ ! -f .venv/bin/activate ]]; then
    echo "ERROR: .venv/bin/activate no encontrado. Entorno virtual roto." >&2
    echo "       Recrea el entorno: uv venv  |  python -m venv .venv" >&2
    exit 1
  fi
  echo "==> Activating virtual environment (.venv)"
  source .venv/bin/activate
  if [[ -z "${VIRTUAL_ENV:-}" ]]; then
    echo "ERROR: Activación de .venv falló. VIRTUAL_ENV no está definido." >&2
    echo "       El entorno puede estar corrupto. Recrea con: uv venv  |  python -m venv .venv" >&2
    exit 1
  fi
fi

echo "==> Syncing dependencies"
"${INSTALL_CMD[@]}"

if [[ "$MODE" == "verify" ]]; then
  echo "==> Running baseline verification"
  set +e
  "${VERIFY_CMD[@]}"
  _RC=$?
  set -e
  if [[ $_RC -ne 0 ]]; then
    # pytest exit 5 = no se recogieron tests. No es fatal en un proyecto sin tests aún.
    if [[ "$PROJECT_TYPE" == "python" && $_RC -eq 5 ]]; then
      echo "WARNING: no se encontraron tests (pytest exit 5). Continuando." >&2
    else
      echo "ERROR: la verificación falló (exit $_RC)." >&2
      exit "$_RC"
    fi
  fi
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
