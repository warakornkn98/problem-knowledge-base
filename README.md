# Problem Knowledge Base

Internal troubleshooting knowledge base for Developer / IT / Infrastructure /
DevOps / Security teams. Record the problems you hit during the day — the error,
the root cause, the fix, the troubleshooting steps and how to prevent a repeat —
then **find them again** when a similar issue shows up.

> The system is not a daily log. It is a searchable knowledge base: when someone
> hits `Connection timed out` they should immediately find the past incident with
> a similar error or context.

---

## Tech stack

| Layer | Choice |
|-------|--------|
| Backend | Go 1.26 · Fiber v2 · GORM · PostgreSQL 16 |
| Frontend | Vue 3 · TypeScript · Vite · Naive UI |
| Search | PostgreSQL Full Text Search (`tsvector` generated column + GIN) + `pg_trgm` fuzzy fallback |
| Auth | JWT (HS256) + bcrypt |
| Infra | Docker · Docker Compose |
| Architecture | Feature-based + Hexagonal (domain / application / infrastructure / handler) |

---

## Quick start (Docker)

```bash
cp .env.example .env          # adjust secrets if you like
docker compose up -d --build
```

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080/api |
| Health | http://localhost:8080/health |
| PostgreSQL | localhost:5432 |

On first boot the backend container runs migrations and seeds baseline data
(`AUTO_MIGRATE=true`, `AUTO_SEED=true`): 11 categories, 14 tags, an example
problem, and an admin account:

```
username: admin
password: admin1234        # from SEED_ADMIN_PASSWORD in .env
```

---

## Local development

### Backend

```bash
cd backend
cp ../.env.example .env        # set DB_HOST=localhost
go run ./cmd/migrate up        # apply migrations
go run ./cmd/seed              # baseline data + admin
go run ./cmd/api               # http://localhost:8080
```

Useful:

```bash
go run ./cmd/migrate status    # which migrations are applied
go test ./...
go vet ./...
```

### Frontend

```bash
cd frontend
npm install
npm run dev                    # http://localhost:5173, proxies /api -> :8080
```

A top-level `Makefile` wraps all of this — run `make help`.

---

## Architecture

```
backend/internal/
├── config/                 env loading
├── server/                 composition root — wires repos → services → handlers
├── shared/                 authx · database · httpx · validatorx · logger · slug
├── user/                   auth (register / login / me)
├── category/               CRUD
├── tag/                    CRUD + resolve-by-name
├── problem/                aggregate: problem + steps + related + similar
│   ├── domain/             entities, enums, ports (interfaces) — no framework
│   ├── application/        use cases / business rules
│   ├── infrastructure/     GORM adapters, the search & facet SQL
│   └── handler/            Fiber transport, request/response mapping
├── search/                 unified search endpoint (facets + relevance)
└── dashboard/              read-model aggregates
```

Rules:

- Handlers contain **no** business logic — parse, call a service, map the result.
- Business logic lives in `application` / `domain`.
- Database access lives in `infrastructure`, behind a port defined in `domain`.
- Features never import each other's `application` package; cross-feature needs
  go through a small port + adapter (see `server/adapters.go`).

---

## Database schema

```
users ──1:N──▶ problems ◀──N:1── problem_categories
                  │
                  ├──1:N──▶ problem_steps
                  ├──M:N──▶ tags            (problem_tags)
                  └──self M:N──▶ problems   (problem_related)
```

Full text search: `problems.search_vector` is a `STORED GENERATED` column
combining title (weight A), error message + tags (B), description / root cause /
solution (C) and prevention (D). Tag text reaches it through the denormalised
`problems.tags_cached` column, kept in sync by the repository on every tag write.
Queries use `websearch_to_tsquery` for ranking with an `ILIKE` / trigram fallback
so misremembered wording still matches.

Migrations are plain SQL in `backend/migrations/`, embedded into the binary and
tracked in `schema_migrations`.

---

## API

Base path `/api`. Every response uses one envelope:

```json
{ "success": true, "data": { }, "message": null }
```

Lists wrap `data` as `{ "items": [], "pagination": { "page", "limit", "total", "total_pages" } }`.

| Area | Endpoints |
|------|-----------|
| Auth | `POST /auth/register` · `POST /auth/login` · `GET /auth/me` |
| Meta | `GET /meta/enums` |
| Problems | `GET /problems` · `POST /problems` · `GET/PUT/DELETE /problems/:id` · `PATCH /problems/:id/status` |
| Search | `GET /problems/search?q=` · `GET /search?q=` (with facets) · `GET /problems/:id/similar` |
| Steps | `GET/POST /problems/:id/steps` · `PUT/DELETE /problems/:id/steps/:stepId` · `PUT /problems/:id/steps/reorder` |
| Related | `GET/POST /problems/:id/related` · `DELETE /problems/:id/related/:relatedId` |
| Tags | `GET /tags` · `POST /tags` · `PUT/DELETE /tags/:id` |
| Categories | `GET /categories` · `POST /categories` · `PUT/DELETE /categories/:id` |
| Dashboard | `GET /dashboard/summary` · `/categories` · `/projects` · `/common-errors` |

`GET /problems` filters (all repeatable, comma-lists accepted):
`q, project, category_id, environment, severity, status, tag,
created_from, created_to, solved_from, solved_to, sort, page, limit`.

All endpoints except `/auth/*`, `/meta/*` and `/health` require
`Authorization: Bearer <token>`.

---

## Development roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| 1 | Project setup, DB, migrations, models, CRUD problem/category/tag | ✅ done |
| 2 | Steps, search, filter, pagination, sorting | ✅ done |
| 2 | Auth (JWT + users) — pulled forward from Phase 4 | ✅ done |
| 3 | Dashboard, similar problems, related problems | ✅ done |
| 3 | Frontend — Vue 3 + Naive UI (dashboard, problems, detail, form, search, categories, tags) | ✅ done |
| 4 | User management, audit log (`updated_by` already captured) | ⏳ |
| 5 | Advanced full-text tuning, Thai dictionary, KB improvements | ⏳ |

## Screens

- **Dashboard** — counters, most-common categories & projects, recurring errors
- **Problems** — data table with server-side search / filter / sort / pagination
- **Problem detail** — description, copyable error block, root cause, solution,
  numbered troubleshooting steps, prevention; sidebar with tags, audit info,
  auto-computed *similar problems* and manual *related problems*
- **New / Edit problem** — full form with dynamic, reorderable steps; the minimum
  for a quick capture is title + project + environment + category
- **Search** — dedicated relevance-ranked search with a faceted sidebar
  (status / severity / environment / category / project / tag, with counts)
- **Categories / Tags** — inline CRUD
