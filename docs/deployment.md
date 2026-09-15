# Deployment guide

## Local development (no Docker)

```bash
# 1. Postgres + Redis only
docker compose up -d postgres redis

# 2. Backend
cd backend
cp .env.example .env        # edit DATABASE_URL/REDIS_ADDR to match the ports above
make migrate-up
make run                    # http://localhost:8080

# 3. Frontend
cd frontend
cp .env.example .env.local  # NEXT_PUBLIC_BACKEND_ORIGIN=http://localhost:8080
npm install
npm run dev                 # http://localhost:3000
```

Seed admin login: `admin@bookreader.local` / `ChangeMe123!` (from migration
`000019_seed_reference_data`). **Rotate this immediately outside local dev.**

## Full stack via Docker Compose

```bash
cp .env.example .env
# edit .env: set POSTGRES_PASSWORD, JWT_ACCESS_SECRET (openssl rand -base64 48),
# and PUBLIC_ORIGIN to the domain/port you'll actually access the site on.
docker compose up -d --build
```

This brings up `postgres`, `redis`, a one-shot `migrate` job, `backend`,
`frontend`, and `nginx` (the only container exposed on the host, default port
80). Nginx same-origins the frontend and API, so the browser never makes a
cross-origin request — no CORS preflight, and cookies flow without any
`SameSite` complications.

### ⚠️ `PUBLIC_ORIGIN` must exactly match what the browser sees

The backend's CORS check compares the browser's `Origin` header against
`CORS_ALLOWED_ORIGINS` (derived from `PUBLIC_ORIGIN`). If you expose nginx on
a non-standard port (e.g. you changed `HTTP_PORT` because 80 was already
taken), `PUBLIC_ORIGIN` **must include that port**
(`http://your-host:8090`), or every login/write request will fail with a 403
from the CORS middleware even though the app otherwise "looks" fine. On
standard ports 80/443 this is a non-issue since browsers omit the default
port from `Origin`.

### Ports

| Variable | Default | Notes |
|---|---|---|
| `HTTP_PORT` | 80 | nginx, the only port that needs to be reachable externally |
| `HTTPS_PORT` | 443 | commented out in `docker-compose.yml` until you provide certs |
| `POSTGRES_HOST_PORT` | 5432 | host access for local tooling only (`psql`, GUI clients) |
| `REDIS_HOST_PORT` | 6379 | same, for `redis-cli` |

### TLS

Terminate TLS at nginx: mount a certs directory and uncomment the `listen
443 ssl` block plus the `certs` volume in both `docker-compose.yml` and
`infra/nginx/nginx.conf`. A `certbot` sidecar or an external reverse proxy
(Caddy, Cloudflare Tunnel, a cloud load balancer) both work equally well —
nginx here doesn't assume any particular ACME setup.

### Storage: local disk vs S3

Default is local disk, persisted in the `uploads_data` named volume (also
mounted read-only into nginx so uploaded covers/PDFs are served directly by
nginx, not proxied through the Go process). To switch to S3-compatible
storage, set in `.env`:

```
STORAGE_PROVIDER=s3
S3_BUCKET=...
S3_REGION=...
S3_ENDPOINT=...        # leave empty for real AWS S3; set for MinIO/R2/etc.
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
S3_PUBLIC_URL=https://your-cdn-or-bucket-domain
```

No code change needed — `pkg/storage.Provider` is selected at startup from
`STORAGE_PROVIDER` (`cmd/api/main.go`).

### Database migrations

The `migrate` service runs once at `docker compose up` and exits; `backend`
waits for it (`condition: service_completed_successfully`) before starting.
To apply a new migration to an already-running stack:

```bash
docker compose run --rm migrate -path /migrations -database "postgres://<user>:<pass>@postgres:5432/<db>?sslmode=disable" up
```

### Backups

```bash
# Backup
docker compose exec postgres pg_dump -U postgres bookreader | gzip > backup-$(date +%F).sql.gz

# Restore
gunzip -c backup-2026-01-01.sql.gz | docker compose exec -T postgres psql -U postgres bookreader
```

### Scaling notes

- **Backend** is stateless (JWT auth, Redis-backed rate limiting) — safe to
  run multiple replicas behind nginx/a load balancer.
- **Redis** is required for the rate limiter and cache to behave consistently
  across replicas; don't skip it in a multi-replica deployment.
- **Excel imports** run in-process on whichever backend replica received the
  `commit` request; the SSE progress stream (`GET
  /admin/import/:id/stream`) must be polled from the same replica or through
  a load balancer with sticky sessions / or by having all replicas poll
  `import_logs` in the database instead (the endpoint already falls back to
  reading from Postgres each tick, so this works even without stickiness —
  just confirm your LB doesn't buffer/cache SSE responses).
- **PostgreSQL** connection pool size (`DB_MAX_OPEN_CONNS`) should be tuned
  relative to `replicas × pool size ≤ Postgres max_connections`.
