# Machinery Safety – Backend

The REST API behind [Machinery Safety](https://github.com/zserub/machinerysafety-FE), a web application for tracking safety risks on production lines and the actions taken to fix them with the principles of PDCA.

It handles login and sessions, role-based permissions, risk entries with their full change history, comments, user preferences and the dropdown values (lines, categories, risk levels, statuses) used by the frontend. It also connects the app to Metabase for statistics and can build the analytics dashboard automatically.

Written in Go, with PostgreSQL for data and MinIO for file storage.

## Features

- **Cookie-based sessions** with bcrypt password hashing
- **Role-based access control** – roles map to permissions such as `entries:create` or `statistics:view`; every endpoint checks them
- **Risk entries** – create, list, search, update and delete; short-term and long-term risks can be linked
- **Change history** – status changes and deadline postponements are recorded, the original deadline is kept
- **Comments** on every entry
- **Metabase integration** – embed URLs for dashboards and cards, and an optional dashboard set up on startup (totals, open and overdue risks, resolution rate and time, risk matrices per line and per responsible, monthly burndown)
- **Automatic database migrations** and a first admin account created on first start
- **Health endpoints** for container orchestration

## Running it

### Development stack (Docker)

Runs the backend together with PostgreSQL, MinIO, the [frontend](https://github.com/zserub/machinerysafety-FE) and, optionally, Metabase. The frontend repository must be checked out next to this one as `../machinerysafety-FE`.

```bash
cp .env.example .env     # then set SESSION_KEY: openssl rand -base64 32
make dev-up              # frontend, backend, PostgreSQL, MinIO
make dev-up-full         # the same, plus Metabase
```

| Service | Address |
|---------|---------|
| Frontend | http://localhost:3001 |
| Backend | http://localhost:8080 |
| Metabase | http://localhost:3000 |
| MinIO console | http://localhost:9001 |

The development stack loads demo data: users `admin`, `alice`, `bob`, … all with the password `password`.

Run `make help` to see all commands (logs, rebuilds, database shell, reset).

### Running the server directly

Requirements: Go 1.25+, a PostgreSQL database and a MinIO server.

```bash
cp .env.example .env     # adjust database, MinIO and SESSION_KEY
make run                 # or: go run .
```

```bash
make build               # binary in bin/server
make test                # unit tests
```

### Production

Build the image:

```bash
docker build --build-arg VERSION=1.0.0 -t machinerysafety-backend .
```

For a complete installation (frontend, backend, PostgreSQL, Metabase and MinIO, also for servers without internet access), see the deployment guide in the frontend repository: [deploy/README.md](https://github.com/zserub/machinerysafety-FE/blob/main/deploy/README.md).

## Configuration

All settings are environment variables; [.env.example](.env.example) lists them with comments. The most important ones:

| Variable | Description |
|----------|-------------|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SCHEMA` | PostgreSQL connection |
| `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`, `MINIO_BUCKET` | File storage |
| `SESSION_KEY` | **Required.** 32-byte base64 key for session cookies (`openssl rand -base64 32`) |
| `SESSION_SECURE` | `true` when users reach the app over HTTPS |
| `CORS_ALLOWED_ORIGINS` | Comma-separated frontend addresses |
| `RUN_MIGRATIONS` | Create or upgrade the database tables on startup |
| `INITIAL_ADMIN_USERNAME`, `INITIAL_ADMIN_PASSWORD` | First admin account, created only when the database has no users |
| `METABASE_URL`, `METABASE_PUBLIC_URL`, `METABASE_API_KEY` | Metabase connection for the statistics page |
| `RUN_METABASE_SETUP`, `METABASE_DATABASE_ID` | Create the analytics dashboard on startup |
| `LOG_LEVEL`, `LOG_FORMAT`, `LOG_FILE_PATH` | Logging (`debug`/`info`/`warn`/`error`, `json`/`text`) |

## Roles

| Role | Can |
|------|-----|
| Visitor | View the dashboard, entries and statistics |
| Responsible | Also create and edit entries |
| Manager | Also delete entries and open the supervisor dashboard |
| Admin | Everything, including settings |

Roles and permissions are stored in the database and can be changed there.

## API

All endpoints except health and version require a logged-in session.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth/login` | Log in with username and password |
| `POST` | `/api/auth/logout` | Log out |
| `GET` | `/api/auth/me` | Current user |
| `GET` | `/api/user/permissions` | Current user's permissions |
| `GET` `PATCH` | `/api/user/preferences` | Language, theme and visible table columns |
| `GET` | `/api/metadata` | Dropdown values: risk levels, statuses, categories, lines, expense types, responsibles |
| `GET` `POST` | `/api/entries` | List and search entries, create an entry |
| `GET` `PATCH` `DELETE` | `/api/entries/{id}` | Read, update or delete an entry |
| `GET` `POST` | `/api/entries/{id}/comments` | Read or add comments |
| `GET` | `/api/metabase/config` | Metabase settings for the frontend |
| `GET` | `/api/metabase/dashboards/{id}/embed`, `/api/metabase/cards/{id}/embed` | Embed URLs |
| `GET` | `/health`, `/health/live`, `/health/ready` | Health checks |
| `GET` | `/api/version` | Running version |

## Built with

[Go](https://go.dev) · [chi](https://github.com/go-chi/chi) · [pgx](https://github.com/jackc/pgx) · [golang-migrate](https://github.com/golang-migrate/migrate) · [sqload](https://github.com/midir99/sqload) · [PostgreSQL](https://www.postgresql.org) · [MinIO](https://min.io) · [Metabase](https://www.metabase.com)

## License

Copyright 2026 Ruben Zsédely

Licensed under the [Apache License, Version 2.0](LICENSE). You are free to use, modify and distribute this software, including commercially, provided that you keep the copyright and [NOTICE](NOTICE) attribution to the original author and mark any files you change.
