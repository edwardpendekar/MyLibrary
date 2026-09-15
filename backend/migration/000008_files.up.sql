-- Generic registry for every uploaded binary (covers, PDFs, import source files).
-- Storage is abstracted: `provider` + `path` let the same row describe a local-disk file
-- today and an S3 object tomorrow without changing calling code.
CREATE TABLE files (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uploaded_by      BIGINT REFERENCES users (id) ON DELETE SET NULL,
    original_name    VARCHAR(255) NOT NULL,
    stored_path      TEXT         NOT NULL,
    provider         VARCHAR(20)  NOT NULL DEFAULT 'local',
    mime_type        VARCHAR(150) NOT NULL,
    size_bytes       BIGINT       NOT NULL DEFAULT 0,
    checksum_sha256  VARCHAR(64),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_files_provider CHECK (provider IN ('local', 's3'))
);

CREATE INDEX idx_files_uploaded_by ON files (uploaded_by);

COMMENT ON TABLE files IS 'Storage-agnostic file registry; the storage.Provider interface resolves stored_path.';
