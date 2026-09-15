CREATE TABLE users (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_id           BIGINT NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    name              VARCHAR(150)        NOT NULL,
    email             CITEXT              NOT NULL,
    password_hash     VARCHAR(255)        NOT NULL,
    avatar_path       TEXT,
    is_active         BOOLEAN             NOT NULL DEFAULT true,
    email_verified_at TIMESTAMPTZ,
    last_login_at     TIMESTAMPTZ,
    created_at        TIMESTAMPTZ         NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ         NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT uq_users_email UNIQUE (email)
);

CREATE INDEX idx_users_role_id ON users (role_id);
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;

COMMENT ON TABLE users IS 'Application accounts (admin/editor/user roles). Guests are unauthenticated.';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete marker; NULL = active account.';
