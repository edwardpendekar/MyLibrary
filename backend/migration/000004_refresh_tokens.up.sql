CREATE TABLE refresh_tokens (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash     VARCHAR(255) NOT NULL,
    replaced_by_id BIGINT       REFERENCES refresh_tokens (id) ON DELETE SET NULL,
    created_by_ip  INET,
    user_agent     TEXT,
    expires_at     TIMESTAMPTZ  NOT NULL,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_refresh_tokens_hash UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

COMMENT ON TABLE refresh_tokens IS 'Rotating refresh tokens (hashed at rest) backing JWT access-token renewal.';
