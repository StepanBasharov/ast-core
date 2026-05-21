# Agent Instructions

## Code change policy

The agent **MUST NOT** write or edit any source files without explicit user confirmation when the change affects existing working code.

Before making any code change the agent must:

1. Show the proposed change (diff or explanation)
2. Explain why it is needed
3. Wait for explicit user confirmation

## Project overview

- **Module:** `backend` (Go 1.26.3)
- **What:** ATS (Applicant Tracking System) backend, early stage
- **IDE:** GoLand (`.idea/` present)

## Go toolchain quirk

`go.mod` requires Go 1.26.3. Local Go may be older. Prefix commands with `GOTOOLCHAIN=auto` when needed:

```
GOTOOLCHAIN=auto go build ./...
```

## Real entrypoint

`cmd/ats-backend/ats-backend.go` — `package main`, wires all dependencies.

`main.go` in the repo root is a stale placeholder and should be deleted. Do not treat it as the entrypoint.

## Layout

```
cmd/ats-backend/          # entrypoint: wires adapters, use-cases, HTTP server
docker/                   # Dockerfile + docker-compose.yml (app + postgres:17)
internal/
  adapters/
    claude/               # AI adapter: Claude tool-use → fills domain.CV
    pdfreader/            # PDF → plain text via ledongthuc/pdf
    postgres/             # CVRepository impl (pgxpool); query constants in query.go
  domain/                 # CV, Skill entities
  transfer/gin/           # HTTP layer: Gin router + handlers
    handlers/
  use-case/               # UploadCVUseCase (upload_cv.go)
    interfaces/           # port interfaces: CVRepository, CVReader, Agent
migrations/               # plain .sql files, applied via docker-entrypoint-initdb.d
pkg/log/                  # Logger interface + zerolog impl (NewZerologLogger)
```

## Architecture

Clean architecture: `domain` ← `use-case` ← `adapters`/`transfer`. Use-cases depend only on interfaces in `internal/use-case/interfaces/`, never on concrete adapters.

## Conventions

- Package names: `usecase` (not `use_case`), `gintransfer` (not `gintranfer`)
- File names: `snake_case.go` (not `UploadNewCV.go`)
- Logger: always use `pkg/log.Logger` interface; log directly via `logger.Error(msg, log.FieldLogger{...})` — no helper wrappers
- Domain IDs: `uuid.UUID` from `github.com/google/uuid`; UUIDs generated in Go (not DB side) on `Create`
- Skills: many-to-many via `skills` + `cv_skills` join table; `domain.Skill{Name: name}` — UUID filled by postgres adapter after upsert

## Database

- Tables: `users_cv`, `skills`, `cv_skills` (see `migrations/001_create_tables.sql`)
- Driver: `github.com/jackc/pgx/v5/pgxpool`
- All skill writes go through `upsertAndLinkSkills` inside a transaction
- `ON CONFLICT (name) DO NOTHING` + separate `SELECT` in `upsertAndLinkSkills` — known race condition under concurrency, not yet fixed

## Docker

```
cd docker
ANTHROPIC_API_KEY=sk-... docker compose up --build
```

Migrations run automatically on first postgres start via `docker-entrypoint-initdb.d`.

## Known open issues

- `main.go` (root) is a dead placeholder — conflicts with `cmd/ats-backend/` during `go build ./...`; delete it
- `updated_at` column exists in `users_cv` but `UpdateCVQuery` does not set it
- `upsertAndLinkSkills`: INSERT + SELECT not atomic — replace with `INSERT ... ON CONFLICT DO UPDATE ... RETURNING uuid`
- `context.Background()` used for DB pool init — no timeout if postgres is unreachable at startup
