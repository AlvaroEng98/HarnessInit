#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

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

PROJECT_TYPE="$(_detect_project_type)"

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

FRAMEWORK="${_FW:-none}"

# Escribe estado detectado solo si no fue provisto por harness-init init
if [[ ! -f .harness-state ]]; then
  printf 'PROJECT_TYPE=%s\nFRAMEWORK=%s\n' "$PROJECT_TYPE" "$FRAMEWORK" > .harness-state
fi

# ---------------------------------------------------------------------------
# Ejecución
# ---------------------------------------------------------------------------

echo "==> Tipo detectado: $PROJECT_TYPE"
echo "==> Framework detectado: $FRAMEWORK"
echo "==> Working directory: $PWD"
echo "==> Syncing dependencies"
"${INSTALL_CMD[@]}"

echo "==> Running baseline verification"
"${VERIFY_CMD[@]}"

echo "==> Startup command"
printf '    %q' "${START_CMD[@]}"
printf '\n'

if [ "${RUN_START_COMMAND:-0}" = "1" ]; then
  echo "==> Starting the app"
  exec "${START_CMD[@]}"
fi

echo "Set RUN_START_COMMAND=1 if you want init.sh to launch the app directly."
