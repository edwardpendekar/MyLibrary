# Authentication flow

JWT access tokens + rotating opaque refresh tokens, delivered as httpOnly
cookies. See `backend/internal/service/auth_service.go`,
`backend/internal/handler/public/auth_handler.go`, and
`frontend/hooks/use-auth.ts`.

```mermaid
sequenceDiagram
    actor User
    participant FE as Next.js (browser)
    participant API as Go API

    User->>FE: submit login form
    FE->>API: POST /api/v1/auth/login {email, password}
    API->>API: bcrypt.CompareHashAndPassword
    API->>API: issue access JWT (15m) + opaque refresh token (30d)
    API->>API: store SHA-256(refresh token) in refresh_tokens
    API-->>FE: Set-Cookie access_token, refresh_token, csrf_token
    FE->>FE: update auth store, redirect

    Note over FE,API: every subsequent request
    FE->>API: fetch(..., credentials: "include", X-CSRF-Token: <csrf cookie>)
    API->>API: middleware.Auth parses access_token JWT
    API->>API: middleware.CSRF compares header to csrf_token cookie

    Note over FE,API: access token expiring
    FE->>API: POST /api/v1/auth/refresh (refresh_token cookie only)
    API->>API: look up hash, check not expired/revoked
    API->>API: issue new access+refresh, revoke old refresh (rotation)
    API-->>FE: new Set-Cookie triplet
```

## Why this shape

- **Access tokens are JWTs** (stateless, cheap to verify on every request) but
  **refresh tokens are opaque random strings**, not JWTs — their hash is
  checked against the `refresh_tokens` table on every use, which is what lets
  the server actually *revoke* a session (a JWT can't be un-issued before it
  expires).
- **Rotation.** Every refresh revokes the old refresh token and issues a new
  one (`replaced_by_id` links them), so a stolen-and-reused refresh token is
  detectable after the legitimate client's next refresh.
- **httpOnly, SameSite=Strict cookies**, not `localStorage` — inaccessible to
  JS, which closes the most common XSS-driven token-theft vector.
- **CSRF double-submit.** Because the cookies are httpOnly, a separate
  JS-readable `csrf_token` cookie is set alongside them; every non-GET request
  must echo it back as `X-CSRF-Token`. Combined with `SameSite=Strict`, this
  covers both modern and legacy-browser CSRF vectors (`pkg/csrf`,
  `internal/middleware/csrf.go`).
- **RBAC.** The JWT carries a `role` claim (`admin`/`editor`/`user`/`guest`);
  `middleware.RequireRole(...)` gates admin routes. The frontend's
  `/admin/layout.tsx` also re-checks the role server-side via `GET /me`
  before rendering, rather than trusting client-side route guards alone.
