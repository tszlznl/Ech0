# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Ech0 is a self-hosted personal microblog (timeline) platform. It is shipped as a single Go binary that serves both the REST API and the built SPA. Backend is Go 1.26+ (Gin + Wire DI + GORM + SQLite via CGO), frontend is Vue 3 + Vite + TypeScript + UnoCSS under `web/`.

## Common commands

All backend commands run from the repository root; frontend commands run from `web/` (or via the `make web-*` wrappers).

```bash
# Backend
make run             # go run ./cmd/ech0 serve (blocks on :6277)
make dev             # Air hot-reload (auto-installs Air via `make air-install` if missing)
make test            # go test ./...
make lint            # golangci-lint run
make fmt             # golangci-lint fmt
make wire            # regenerate internal/di/wire_gen.go (run after changing provider sets / DI graph)
make wire-check      # fails if wire_gen.go is stale vs. wire.go
make swagger         # swag init -g internal/server/server.go -o internal/swagger --parseInternal

# Frontend (from web/)
pnpm install
pnpm dev             # Vite dev server on :5173, proxies to backend on :6277
pnpm build           # type-check + vite build
pnpm test:unit       # vitest run
pnpm lint            # eslint . --fix
pnpm format          # prettier --write src/
pnpm i18n:check      # runs key / unused / hardcoded / pseudo-smoke checks (required before PR)

# Full pre-PR verification (mandatory per CONTRIBUTING.md)
make check           # alias of make dev-lint: backend fmt+lint+swagger + web format+lint+i18n:check

# Single Go test
go test ./internal/middleware -run TestAuth     # example
go test -run TestName ./path/to/pkg             # by name
```

Run a single frontend test: `pnpm -C web exec vitest run path/to/file.spec.ts` (or `-t "test name"`).

Binary entrypoint is `cmd/ech0/main.go`. CLI verbs (Cobra): `ech0 serve` (HTTP), bare `ech0`/`ech0 tui` (TUI), `ech0 version`, `ech0 info`, `ech0 hello`. The `backup.go` CLI command provides snapshot export outside the web UI.

## Architecture

### Layered backend with Wire DI

Backend follows a strict layered architecture — **handler → service → repository → database** — with Google Wire generating the dependency graph. Each domain (echo, user, auth, comment, connect, file, setting, dashboard, agent, backup, migration, init, common) has parallel packages under `internal/handler/<x>`, `internal/service/<x>`, `internal/repository/<x>`, and `internal/model/<x>`.

- `internal/di/wire.go` declares provider sets (`HandlerSet`, `EventSet`, `TaskerSet`, `MigratorSet`, `MiddlewareSet`, `InfraSet`, `RuntimeSet`) and the `BuildApp` injector that composes the full runtime. **If you add/remove a constructor or change a binding, run `make wire`** before committing.
- Cross-domain aliases are required when importing layers: `xxxHandler`, `xxxService`, `xxxRepository`, `xxxModel`, `xxxUtil` (enforced by existing code; see README "Start Backend & Frontend" note).
- `internal/app` is a generic component lifecycle orchestrator. `internal/server` is the thin Gin/HTTP `Component` it manages. Other `Component`s (Tasker, migrator, event registrar) are started/stopped alongside the HTTP server.
- `internal/bootstrap/bootstrap.go` runs before Cobra dispatches: loads config, initializes the zap-based logger, sets host env defaults. Config is accessed via `config.Config()` (singleton).
- HTTP routes live in `internal/router/*.go`, registered per domain and wired up in `internal/server/provider.go`. Swagger annotations on handlers drive `internal/swagger/` output.

### Event bus (Busen)

Ech0 uses the in-repo **Busen** library (`github.com/lin-snow/Busen`) as an async in-process event bus. Publishers live at `internal/event/publisher`, subscribers at `internal/event/subscriber` (agent processor, backup scheduler, dead-letter resolver), contracts at `internal/event/contracts`, wiring at `internal/event/registry`. The bus decouples comment/echo/user events from side effects like webhooks, agent runs, and backups. Runtime tuning is via `ECH0_EVENT_*` env vars (buffers, parallelism, webhook worker pool) — see README "Event Runtime Parameters".

Webhook dispatch (`internal/webhook`) and agent processing (`internal/agent`) are implemented as event subscribers, not as inline handler calls. When adding cross-cutting side effects, prefer publishing an event over invoking services directly from handlers.

### Storage (VireFS)

`internal/storage` is a unified abstraction over local disk and S3-compatible object stores. Files are addressed by a flat `key`; `schema.Resolve` + `PathPrefix` map keys to on-disk paths or S3 object keys. Stored `File.url` is a snapshot of the UI-visible URL at write time. The `/api/files` static route serves local content; the `stream` routes are authenticated. `S3SettingStore` is bound to `KeyValueRepository` so S3 config lives in the settings DB, not env. Switching providers / migrating between local and S3 is documented at `docs/usage/storage-migration.md`.

### Frontend

- Vue 3 SFCs in `web/src`, Pinia stores, Vue Router, i18n via `vue-i18n`, UnoCSS (Wind4 preset), markdown via `markdown-it` + Vditor editor.
- i18n guardrails in `web/scripts/` (key completeness, unused keys, hardcoded strings, pseudo-locale smoke) are part of `make check` — **do not introduce hardcoded UI strings**; use translation keys.
- Vite serves `:5173` during dev and proxies `/api` to the backend on `:6277`. For production, the backend embeds the built SPA (see `template/` and `internal/handler/web`).

### Configuration

`internal/config/config.go` is the single config source; env vars are parsed via `caarlos0/env`. See `.env.example` for the full set (JWT secret, server port, DB path, log, S3, event runtime, etc.). Defaults target `./data/` for SQLite + uploaded files; Docker images mount `/app/data`.

## Conventions to respect

- **Before a PR**: `make check` is mandatory (enforces backend lint, frontend lint, i18n checks). `go build ./...` and `pnpm build` must pass. Regenerate Swagger (`make swagger`) whenever routes or request/response shapes change and commit `internal/swagger/`.
- **DI changes**: regenerate with `make wire`; CI runs `make wire-check`.
- **v3 → v4 migration is not in-place**: export a snapshot in v3 panel, redeploy v4, then use "v3 Migration" in the v4 panel. Migration code lives in `internal/migrator` and `internal/service/migrator`.
- **Integration comment endpoint**: `POST /api/comments/integration` intentionally bypasses captcha/form-token — it requires an access token with `comment:write` scope and `integration` audience. Preserve this behavior.
- **Access tokens**: scope/audience/`typ` design is documented at `docs/dev/access-token-scope-design.md`; implementation is authoritative.
- **Layered import aliases** (required): `xxxHandler`, `xxxService`, `xxxRepository`, `xxxModel`, `xxxUtil`.
- **Logging**: use the project zap wrapper at `internal/util/log` with a `module` field; see `docs/dev/logging.md` for field conventions.

## Useful reference docs in-repo

- `docs/dev/auth-design.md`, `docs/dev/access-token-scope-design.md` — auth model & token scopes
- `docs/dev/i18n-contract.md` — frontend/backend i18n contract (locale header, error field shapes, key naming)
- `docs/dev/table-design-standard.md` — admin panel table component conventions
- `docs/dev/logging.md`, `docs/dev/timezone-design.md`, `docs/dev/table-design-standard.md`
- `docs/usage/storage-migration.md`, `docs/usage/mcp-usage.md`, `docs/usage/webhook-usage.md`
- `CONTRIBUTING.md` — PR workflow and pre-submission checks
