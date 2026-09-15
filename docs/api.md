# API reference

The full, always-current reference is generated from code annotations
(`@Summary`/`@Router`/... comments on each handler in
`backend/internal/handler/**`) via [swaggo](https://github.com/swaggo/swag).

- **Interactive Swagger UI**: `http://<backend-host>/swagger/index.html`
  (non-production only — gated off when `APP_ENV=production`, see
  `internal/router/router.go`).
- **Raw OpenAPI 2.0 spec**: `backend/docs/swagger.json` /
  `backend/docs/swagger.yaml` (regenerate with `make swagger` after changing
  any handler's annotations).

## Envelope

Every response — success or error — uses the same shape
(`pkg/response/response.go`):

```json
{ "success": true, "data": { ... }, "meta": { "next_cursor": "...", "has_more": true, "limit": 20 } }
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "...", "fields": { "email": "is required" } } }
```

`meta` is only present on cursor-paginated list endpoints. `error.fields` is
only present for `VALIDATION_ERROR`.

## Endpoint groups

| Prefix | Auth | Purpose |
|---|---|---|
| `POST /api/v1/auth/*` | none → sets cookies | register, login, refresh, logout |
| `GET /api/v1/me*` | cookie | current user, favorites |
| `GET /api/v1/books*` | optional | list/detail/chapters/verses (public, published only) |
| `GET /api/v1/search/*` | optional | full-text search (verses, books) |
| `POST/DELETE /api/v1/books/:id/favorite` | cookie | toggle favorite |
| `* /api/v1/bookmarks*`, `/notes*` | cookie | per-user bookmarks/notes |
| `* /api/v1/admin/books*` | admin/editor | CRUD (any status), cover/PDF upload |
| `* /api/v1/admin/import*` | admin/editor | upload, preview, commit, SSE progress, history |
| `* /api/v1/admin/languages*`, `/categories*` | admin/editor | reference data CRUD |
| `* /api/v1/admin/users*` | admin | role changes, deactivation |
| `GET /api/v1/admin/stats` | admin/editor | dashboard counters |
| `GET /api/v1/admin/audit-logs` | admin | audit trail |

## Pagination

List endpoints use opaque cursor pagination (`pkg/pagination`), never
`offset`/`page` — pass the previous response's `meta.next_cursor` as
`?cursor=...` to fetch the next page. This keeps deep pages O(1) regardless
of how many rows precede them, which matters once `verses` holds 1M+ rows.

## Authentication for non-browser clients

Cookie-based auth (`credentials: "include"`) is what the frontend uses. A
non-browser client (a script, Postman) can instead send `Authorization:
Bearer <access_token>` — `middleware.Auth` accepts either. Note that the
CSRF check only applies to cookie-authenticated requests, so a
bearer-token client does not need `X-CSRF-Token`.
