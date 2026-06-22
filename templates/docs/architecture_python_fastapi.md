# Arquitectura: Python / FastAPI

## Estructura de directorios

```
├── app/                          # Paquete principal de la aplicación
│   ├── __init__.py
│   ├── main.py                   # Entry point FastAPI (crea la app, registra routers)
│   ├── logging_config.py         # Configuración centralizada de logging (una sola puerta)
│   ├── config.py                 # Settings vía Pydantic BaseSettings + .env
│   ├── dependencies.py           # Inyección de dependencias (FastAPI built-ins, no frameworks)
│   │
│   ├── routers/                  # Capa HTTP — solo enrutan, no contienen lógica
│   │   ├── __init__.py
│   │   └── users.py              # Endpoints /api/v1/users (GET, POST, PUT, DELETE)
│   │
│   ├── services/                 # Capa de negocio — "business seam"
│   │   ├── __init__.py
│   │   └── user_service.py       # UserService: lógica pura, sin dependencias HTTP
│   │
│   ├── repositories/             # Capa de acceso a datos (Repository Pattern)
│   │   ├── __init__.py
│   │   └── user_repository.py    # CRUD contra DB; aislable para tests
│   │
│   ├── models/                   # Modelos ORM (SQLAlchemy / SQLModel)
│   │   ├── __init__.py
│   │   └── user.py
│   │
│   └── schemas/                  # Pydantic schemas — request / response
│       ├── __init__.py
│       └── user.py               # UserCreate, UserRead, UserUpdate
│
├── tests/                        # Tests rápidos y herméticos
│   ├── __init__.py
│   ├── conftest.py               # Fixtures compartidos (TestClient, DB en memoria)
│   └── test_user_service.py      # Unit tests del servicio (sin I/O real)
│
├── .env                          # Variables de entorno locales (no se versiona)
├── .env.example                  # Plantilla pública de variables de entorno
├── .python-version               # Versión exacta de Python (compatible con uv/pyenv)
├── .gitignore
├── pyproject.toml                # Dependencias + config de Ruff, Mypy, pytest (todo en uno)
├── Dockerfile                    # Imagen de producción (multi-stage, efficient layer caching)
├── docker-compose.yml            # Entorno local completo (app + DB + extras)
└── README.md
```

## Capas y responsabilidades

| Capa | Responsabilidad | Prohibido |
|------|----------------|-----------|
| `routers/` | Enrutar HTTP → service. Validar con schemas Pydantic. | Lógica de negocio, acceso directo a DB |
| `services/` | Lógica pura de negocio. Sin estado HTTP. | Dependencias HTTP, ORM directo |
| `repositories/` | CRUD contra DB. Única capa que toca ORM. | Lógica de negocio |
| `models/` | Modelos ORM (SQLAlchemy/SQLModel). | Schemas Pydantic |
| `schemas/` | Contratos request/response Pydantic. | Modelos ORM |
| `dependencies.py` | Inyección de dependencias FastAPI. | Lógica de negocio |
| `config.py` | Settings via `BaseSettings`. Lee `.env`. | Side effects en import |
| `logging_config.py` | Única puerta de configuración de logging. | Configurar logging en otros módulos |

## Convenciones

- `pyproject.toml` como fuente única de configuración: dependencias, Ruff, Mypy, pytest.
- `uv` como dependency manager preferido (`uv sync` / `uv add`).
- `conftest.py`: fixtures con `TestClient` y DB en memoria (sin I/O real en unit tests).
- `.env` no se versiona; `.env.example` sí — documenta todas las variables requeridas.
- Dockerfile multi-stage: stage `builder` instala deps, stage `runtime` copia solo lo necesario.
- Versión de Python fijada en `.python-version` (compatible con `uv` y `pyenv`).
- Prefijo de rutas: `/api/v1/` por defecto.
