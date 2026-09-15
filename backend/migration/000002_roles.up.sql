CREATE TABLE roles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(32)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_roles_name UNIQUE (name)
);

COMMENT ON TABLE roles IS 'Static RBAC roles: admin, editor, user, guest.';
