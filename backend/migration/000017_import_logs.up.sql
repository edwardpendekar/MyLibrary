CREATE TABLE import_logs (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uploaded_by           BIGINT REFERENCES users (id) ON DELETE SET NULL,
    source_file_id        BIGINT REFERENCES files (id) ON DELETE SET NULL,
    filename              VARCHAR(255) NOT NULL,
    status                VARCHAR(20)  NOT NULL DEFAULT 'pending',
    mode                  VARCHAR(20)  NOT NULL DEFAULT 'insert',
    total_rows            INTEGER      NOT NULL DEFAULT 0,
    processed_rows        INTEGER      NOT NULL DEFAULT 0,
    books_created         INTEGER      NOT NULL DEFAULT 0,
    chapters_created      INTEGER      NOT NULL DEFAULT 0,
    sections_created      INTEGER      NOT NULL DEFAULT 0,
    verses_inserted       INTEGER      NOT NULL DEFAULT 0,
    verses_updated        INTEGER      NOT NULL DEFAULT 0,
    verses_skipped        INTEGER      NOT NULL DEFAULT 0,
    error_message         TEXT,
    started_at            TIMESTAMPTZ,
    finished_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_import_logs_status CHECK (
        status IN ('pending', 'validating', 'ready', 'importing', 'completed', 'failed', 'rolled_back')
    ),
    CONSTRAINT chk_import_logs_mode CHECK (mode IN ('insert', 'upsert'))
);

CREATE INDEX idx_import_logs_uploaded_by ON import_logs (uploaded_by);
CREATE INDEX idx_import_logs_status ON import_logs (status);
CREATE INDEX idx_import_logs_created_at ON import_logs (created_at DESC);

COMMENT ON TABLE import_logs IS 'One row per Excel/CSV import job; polled/streamed by the admin UI progress bar.';
