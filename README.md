# Book Reader Platform

A scalable, production-oriented platform for reading structured, multi-language text corpora
(Book → Chapter → Verse — e.g. Bible-style texts) with a PDF reader, full-text search, and an
admin console for bulk Excel import.

## Stack

| Layer      | Technology |
|------------|------------|
| Frontend   | Next.js 15, React 19, TypeScript, TailwindCSS, shadcn/ui, TanStack Query, Zustand, React Hook Form, Zod, next-intl |
| Backend    | Go 1.25+, Gin, GORM, PostgreSQL 17, Redis, JWT, Clean Architecture |
| Database   | PostgreSQL 17 (+ full text search) |
| Cache      | Redis |
| Storage    | Local disk (S3-compatible interface, swappable) |
| PDF        | PDF.js |
| Deployment | Docker, Docker Compose, Nginx |

## Repository layout

```
backend/    Go API (clean architecture: domain / repository / service / handler)
frontend/   Next.js app (App Router)
infra/      nginx, postgres, redis configuration for docker-compose
docs/       ERD, flow diagrams, API & deployment docs
scripts/    dev/ops helper scripts (sample data generator, seeders)
```

## Getting started (local development)

See [docs/deployment.md](docs/deployment.md) for full instructions. Quick start:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
docker compose up -d postgres redis
cd backend && make migrate-up && make run
cd frontend && npm install && npm run dev
```

Or run everything containerized:

```bash
docker compose up --build
```

### Testing

```bash
cd backend
go test ./...                              # unit tests (fakes, no external services)
go test -tags=integration ./test/integration/...  # requires a running Docker daemon;
                                                    # spins up throwaway Postgres/Redis containers
```

The integration suite exercises the Excel/CSV import pipeline and full-text search directly against real Postgres (generated `tsvector` columns, `COPY`-based batch upsert) rather than fakes — see [test/integration/helpers_test.go](backend/test/integration/helpers_test.go).

## Documentation

- [Architecture & ERD](docs/architecture.md)
- [Import flow](docs/flow-import.md)
- [Auth flow](docs/flow-auth.md)
- [Reader flow](docs/flow-reader.md)
- [API reference](docs/api.md) (see also `/swagger` on the running backend)
- [Deployment guide](docs/deployment.md)
