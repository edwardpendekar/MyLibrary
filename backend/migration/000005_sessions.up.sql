-- Lightweight session/device tracking, independent from refresh-token rotation.
-- Lets a user see "active sessions" and lets an admin/audit trail see who is logged in from where.
CREATE TABLE sessions (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token_id BIGINT       REFERENCES refresh_tokens (id) ON DELETE SET NULL,
    ip_address       INET,
    user_agent       TEXT,
    last_active_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    expires_at       TIMESTAMPTZ  NOT NULL,
    revoked_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
