# Tasker

Tasker is a task-management backend and supporting TypeScript workspace for API contracts, OpenAPI generation, and transactional email templates. The backend is written in Go with Echo, PostgreSQL, Redis-backed background jobs, Clerk authentication, AWS S3 attachments, Resend email delivery, and New Relic observability.

## Table of contents

- [Features](#features)
- [Repository structure](#repository-structure)
- [Tech stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Running the backend](#running-the-backend)
- [Database migrations](#database-migrations)
- [API overview](#api-overview)
- [OpenAPI documentation](#openapi-documentation)
- [Background jobs](#background-jobs)
- [Email templates](#email-templates)
- [Workspace packages](#workspace-packages)
- [Development commands](#development-commands)
- [Testing and quality checks](#testing-and-quality-checks)
- [Notes and caveats](#notes-and-caveats)

## Features

- Authenticated todo API protected by Clerk bearer-token authentication.
- Todo CRUD with status, priority, due date, metadata, categories, subtasks, comments, and attachments.
- Category management scoped per user.
- Comment creation, listing, update, and deletion.
- File attachment upload/download flow backed by AWS S3 presigned URLs.
- PostgreSQL persistence using `pgx`/`pgxpool`.
- Embedded database migrations using `tern`.
- Redis-backed background processing with `asynq`.
- Scheduled cron jobs for due-date reminders, overdue notifications, weekly reports, and auto-archiving.
- Transactional emails generated with React Email and delivered through Resend.
- OpenAPI generation from shared TypeScript/Zod contracts.
- New Relic tracing/logging instrumentation for HTTP, PostgreSQL, and Redis.
- Global middleware for CORS, security headers, request IDs, request logging, recovery, rate limiting, and tracing.

## Repository structure

```text
.
├── apps/
│   └── backend/                 # Go API server and cron runner
│       ├── cmd/
│       │   ├── tasker/           # HTTP API entry point
│       │   └── cron/             # CLI for scheduled jobs
│       ├── internal/
│       │   ├── config/           # Environment/config loading
│       │   ├── cron/             # Cron job registry and jobs
│       │   ├── database/         # PostgreSQL connection and migrations
│       │   ├── handler/          # Echo HTTP handlers
│       │   ├── lib/              # AWS, email, jobs, utilities
│       │   ├── middleware/       # Auth, tracing, rate limit, global middleware
│       │   ├── model/            # Domain models and DTOs
│       │   ├── repository/       # SQL data access
│       │   ├── router/           # Route registration
│       │   ├── server/           # Server composition
│       │   └── service/          # Business logic
│       ├── static/               # Static OpenAPI UI/assets
│       ├── templates/            # Exported email templates
│       ├── Taskfile.yml          # Backend task automation
│       └── go.mod
├── packages/
│   ├── emails/                   # React Email templates
│   ├── openapi/                  # OpenAPI generation from contracts
│   └── zod/                      # Shared Zod schemas and TypeScript types
├── package.json                  # Bun/Turbo workspace scripts
├── turbo.json                    # Turbo task pipeline
└── bun.lock
```

## Tech stack

### Backend

- Go `1.26.3` as declared in `apps/backend/go.mod`
- Echo v4 HTTP framework
- PostgreSQL with `pgx`/`pgxpool`
- `tern` for SQL migrations
- Redis and `asynq` for background jobs
- Clerk for authentication and user lookup
- AWS SDK v2 for S3 file storage
- Resend for email delivery
- Zerolog for structured logging
- New Relic for observability
- Testcontainers for integration testing support

### TypeScript workspace

- Bun `1.3.3`
- Turbo
- TypeScript
- Zod
- `@ts-rest/*` and `@anatine/zod-openapi` for contract/OpenAPI generation
- React Email for email templates

## Prerequisites

Install the following before running the project locally:

- [Bun](https://bun.sh/) `1.3.3` or compatible
- Node.js `>=32` as declared in the root `package.json`
- Go `1.26.3` or compatible with the module version
- PostgreSQL
- Redis
- [Task](https://taskfile.dev/) for backend Taskfile commands
- `tern` CLI if you want to run migration commands manually from the Taskfile
- AWS credentials and an S3 bucket for attachment uploads
- Clerk secret key for API authentication
- Resend API key for email delivery
- New Relic license key if observability is enabled/configured

## Getting started

Clone the repository and install the TypeScript workspace dependencies:

```sh
bun install
```

Install Go dependencies for the backend:

```sh
cd apps/backend
go mod download
```

Create a local backend environment file from the sample if available:

```sh
cd apps/backend
cp .env.sample .env
```

> The `.env.sample` file exists under `apps/backend`, but its contents may be hidden by editor privacy settings. See [Configuration](#configuration) for the config fields expected by the Go code.

Start PostgreSQL and Redis, then run the backend from `apps/backend`:

```sh
task run
```

or directly:

```sh
go run ./cmd/tasker
```

## Configuration

The backend loads environment variables with the `TASKER_` prefix from the process environment and from `.env` via `godotenv/autoload`.

Configuration is defined in `apps/backend/internal/config/config.go` and `apps/backend/internal/config/observability.go`.

### Required application config

| Config area | Field | Purpose |
| --- | --- | --- |
| `primary` | `env` | Runtime environment. `local` skips automatic embedded migrations on API startup. |
| `server` | `port` | HTTP server port. |
| `server` | `read_timeout` | HTTP read timeout in seconds. |
| `server` | `write_timeout` | HTTP write timeout in seconds. |
| `server` | `idle_timeout` | HTTP idle timeout in seconds. |
| `server` | `cors_allowed_origins` | Allowed CORS origins. |
| `database` | `host` | PostgreSQL host. |
| `database` | `port` | PostgreSQL port. |
| `database` | `user` | PostgreSQL user. |
| `database` | `password` | PostgreSQL password. |
| `database` | `name` | PostgreSQL database name. |
| `database` | `ssl_mode` | PostgreSQL SSL mode, for example `disable` locally. |
| `database` | `max_open_conns` | Maximum open database connections. |
| `database` | `max_idle_conns` | Maximum idle database connections. |
| `database` | `conn_max_life_time` | Connection max lifetime. |
| `database` | `conn_max_idle_time` | Connection max idle time. |
| `auth` | `secret_key` | Clerk secret key. |
| `redis` | `address` | Redis address, for example `localhost:6379`. |
| `integration` | `resend_api_key` | Resend API key. |
| `aws` | `region` | AWS region. |
| `aws` | `access_key_id` | AWS access key ID. |
| `aws` | `secret_access_key` | AWS secret access key. |
| `aws` | `upload_bucket` | S3 bucket used for todo attachments. |

### Optional/defaulted config

| Config area | Field | Default/behavior |
| --- | --- | --- |
| `observability` | `service_name` | Overridden to `tasker`. |
| `observability` | `environment` | Overridden from `primary.env`. |
| `observability.logging` | `level` | Defaults to `info` in default config. Valid values: `debug`, `info`, `warn`, `error`. |
| `observability.logging` | `format` | Defaults to `json`. |
| `observability.logging` | `slow_query_threshold` | Defaults to `100ms`. |
| `observability.new_relic` | `license_key` | Empty by default. Required by New Relic if enabled in your environment. |
| `observability.new_relic` | `appLog_forwarding_enabled` | Defaults to `true`. |
| `observability.new_relic` | `distributed_tracing_enabled` | Defaults to `true`. |
| `observability.new_relic` | `debug_logging` | Defaults to `false`. |
| `observability.health_checks` | `enabled` | Defaults to `true`. |
| `observability.health_checks` | `interval` | Defaults to `30s`. |
| `observability.health_checks` | `timeout` | Defaults to `30s`. |
| `observability.health_checks` | `checks` | Defaults to `database`, `redis`. |
| `cron` | `archive_days_threshold` | Defaults to `30`. |
| `cron` | `batch_size` | Defaults to `100`. |
| `cron` | `reminder_hours` | Defaults to `24`. |
| `cron` | `max_todos_per_user_notification` | Defaults to `10`. |

### Migration DSN

The Taskfile migration commands use a separate `TASKER_DB_DSN` variable:

```sh
TASKER_DB_DSN='postgres://user:password@localhost:5432/tasker?sslmode=disable'
```

## Running the backend

From `apps/backend`:

```sh
task run
```

Equivalent direct Go command:

```sh
go run ./cmd/tasker
```

When the API starts, it:

1. Loads and validates configuration.
2. Initializes New Relic-aware logging.
3. Runs embedded migrations automatically unless `primary.env` is `local`.
4. Connects to PostgreSQL.
5. Connects to Redis, logging an error but continuing if Redis is unavailable.
6. Starts the `asynq` job server.
7. Registers HTTP routes and middleware.
8. Starts the HTTP server on `server.port`.

## Database migrations

Migrations live in `apps/backend/internal/database/migrations`:

- `001_setup.sql` creates shared helper functions such as camel-case JSON conversion and `updated_at` triggers.
- `002_todos.sql` creates categories, todos, comments, indexes, constraints, and triggers.
- `003_attachments.sql` creates todo attachment storage metadata.

Create a new migration:

```sh
cd apps/backend
task migrations:new name=add_some_feature
```

Run migrations manually:

```sh
cd apps/backend
TASKER_DB_DSN='postgres://user:password@localhost:5432/tasker?sslmode=disable' task migrations:up
```

The `migrations:up` task prompts for confirmation before applying migrations.

## API overview

System routes:

| Method | Path | Description | Auth |
| --- | --- | --- | --- |
| `GET` | `/status` | Health check. | No |
| `GET` | `/docs` | OpenAPI UI. | No |
| `GET` | `/static/*` | Static assets, including generated OpenAPI JSON. | No |

Versioned API base path: `/api/v1`

### Todos

All todo routes require authentication.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/todos` | Create a todo. |
| `GET` | `/api/v1/todos` | List todos with pagination, sorting, search, and filters. |
| `GET` | `/api/v1/todos/stats` | Get todo status/overdue counts. |
| `GET` | `/api/v1/todos/:id` | Get a todo by ID. |
| `PATCH` | `/api/v1/todos/:id` | Update a todo. |
| `DELETE` | `/api/v1/todos/:id` | Delete a todo. |

Todo fields include:

- `title`
- `description`
- `status`: `draft`, `active`, `completed`, `archived`
- `priority`: `low`, `medium`, `high`
- `dueDate`
- `parentTodoId`
- `categoryId`
- `metadata` with tags, reminder, color, and difficulty

List query filters include:

- `page`, `limit`
- `sort`: `created_at`, `updated_at`, `title`, `priority`, `due_date`
- `order`: `asc`, `desc`
- `search`
- `status`
- `priority`
- `categoryId`
- `parentTodoId`
- `dueFrom`, `dueTo`
- `overdue`
- `completed`

### Todo comments

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/todos/:id/comments` | Add a comment to a todo. |
| `GET` | `/api/v1/todos/:id/comments` | List comments for a todo. |
| `PATCH` | `/api/v1/comments/:id` | Update a comment. |
| `DELETE` | `/api/v1/comments/:id` | Delete a comment. |

### Todo attachments

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/todos/:id/attachments` | Upload an attachment for a todo. |
| `DELETE` | `/api/v1/todos/:id/attachments/:attachmentId` | Delete an attachment. |
| `GET` | `/api/v1/todos/:id/attachments/:attachmentId/download` | Get a presigned download URL. |

### Categories

All category routes require authentication.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/categories` | Create a category. |
| `GET` | `/api/v1/categories` | List categories with pagination, sorting, and search. |
| `PATCH` | `/api/v1/categories/:id` | Update a category. |
| `DELETE` | `/api/v1/categories/:id` | Delete a category. |

Category fields include:

- `name`
- `color` as a hex color
- `description`

## OpenAPI documentation

OpenAPI-related code lives in `packages/openapi` and shared schemas live in `packages/zod`.

Generate the OpenAPI document:

```sh
cd packages/openapi
bun run gen
```

The generator writes OpenAPI JSON to:

- `packages/openapi/openapi.json`
- `apps/backend/static/openapi.json`

The backend serves API documentation at:

```text
GET /docs
```

The static JSON is available under:

```text
GET /static/openapi.json
```

## Background jobs

The backend uses `asynq` and Redis for background email jobs.

Task types include:

- `email:welcome`
- `email:reminder`
- `email:weekly_report`

The cron runner is available at `apps/backend/cmd/cron`.

List available cron jobs:

```sh
cd apps/backend
go run ./cmd/cron List
```

Run a specific job:

```sh
cd apps/backend
go run ./cmd/cron due-date-reminders
go run ./cmd/cron overdue-notifications
go run ./cmd/cron weekly-reports
go run ./cmd/cron auto-archive
```

Available jobs:

| Job | Description |
| --- | --- |
| `due-date-reminders` | Enqueues email reminders for todos due soon. |
| `overdue-notifications` | Enqueues notifications for overdue todos. |
| `weekly-reports` | Enqueues weekly productivity report emails. |
| `auto-archive` | Archives completed todos older than the configured threshold. |

## Email templates

React Email templates live in `packages/emails/src/templates`:

- `welcome.tsx`
- `due-date-reminder.tsx`
- `overdue-notification.tsx`
- `weekly-reports.tsx`

Preview email templates locally:

```sh
cd packages/emails
bun run dev
```

Export templates into the Go backend template directory:

```sh
cd packages/emails
bun run export
```

The export command writes to:

```text
apps/backend/templates/emails
```

## Workspace packages

### `@tasker/zod`

Shared Zod schemas and TypeScript types for Tasker resources.

Scripts:

```sh
cd packages/zod
bun run build
bun run dev
bun run clean
```

### `@tasker/openapi`

OpenAPI contract generation package built on top of `@tasker/zod`, `@ts-rest/*`, and `@anatine/zod-openapi`.

Scripts:

```sh
cd packages/openapi
bun run gen
bun run build
bun run dev
bun run clean
```

### `@tasker/emails`

React Email templates and export tooling.

Scripts:

```sh
cd packages/emails
bun run dev
bun run export
```

## Development commands

Root commands use Turbo and Bun:

```sh
bun run build
bun run dev
bun run format:check
bun run format:fix
bun run lint
bun run lint:fix
bun run typecheck
bun run clean
```

Backend Taskfile commands from `apps/backend`:

```sh
task help
task run
task migrations:new name=migration_name
task migrations:up
task tidy
```

Go commands from `apps/backend`:

```sh
go test ./...
go fmt ./...
go mod tidy
go mod verify
go run ./cmd/tasker
go run ./cmd/cron List
```

## Testing and quality checks

Run backend tests:

```sh
cd apps/backend
go test ./...
```

Run backend formatting/dependency hygiene:

```sh
cd apps/backend
task tidy
```

Run TypeScript builds/checks through Turbo:

```sh
bun run build
bun run typecheck
```

> Some root scripts depend on individual package scripts. At the time of writing, the root `package.json` workspaces include `packages/*`; the Go backend is managed separately under `apps/backend`.

## Notes and caveats

- API routes under `/api/v1` are authenticated with Clerk via the auth middleware.
- `primary.env=local` skips automatic embedded migrations during API startup. Use `task migrations:up` manually when working locally.
- Redis startup failure is logged but does not stop the HTTP server; background jobs require Redis to function correctly.
- Attachment functionality requires valid AWS credentials and an S3 bucket.
- Email delivery requires a valid Resend API key and reachable email templates.
- OpenAPI JSON should be regenerated after changing schemas or contracts in `packages/zod` or `packages/openapi`.
- The `apps/backend/.env.sample` file is intentionally not reproduced here because it may contain environment details hidden by privacy settings.
