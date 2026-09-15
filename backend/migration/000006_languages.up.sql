CREATE TABLE languages (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code        VARCHAR(10)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    native_name VARCHAR(100) NOT NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_languages_code UNIQUE (code)
);

COMMENT ON TABLE languages IS 'ISO-639-1 style language catalog used for book metadata and UI locale.';
