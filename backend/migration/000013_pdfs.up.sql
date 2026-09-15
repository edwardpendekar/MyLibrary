CREATE TABLE pdfs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    book_id     BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    file_id     BIGINT      NOT NULL REFERENCES files (id) ON DELETE CASCADE,
    page_count  INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_pdfs_book_id UNIQUE (book_id)
);

CREATE INDEX idx_pdfs_file_id ON pdfs (file_id);
