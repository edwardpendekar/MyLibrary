CREATE TABLE books (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug            VARCHAR(255) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    author          VARCHAR(255),
    description     TEXT,
    language_id     BIGINT REFERENCES languages (id) ON DELETE SET NULL,
    category_id     BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    year            SMALLINT,
    isbn            VARCHAR(32),
    cover_path      TEXT,
    cover_file_id   BIGINT REFERENCES files (id) ON DELETE SET NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    chapters_count  INTEGER      NOT NULL DEFAULT 0,
    verses_count    INTEGER      NOT NULL DEFAULT 0,
    view_count      BIGINT       NOT NULL DEFAULT 0,
    created_by      BIGINT REFERENCES users (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    search          tsvector GENERATED ALWAYS AS (
                        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
                        setweight(to_tsvector('simple', coalesce(author, '')), 'B') ||
                        setweight(to_tsvector('simple', coalesce(isbn, '')), 'C') ||
                        setweight(to_tsvector('english', coalesce(description, '')), 'D')
                     ) STORED,
    CONSTRAINT uq_books_slug UNIQUE (slug),
    CONSTRAINT uq_books_isbn UNIQUE (isbn),
    CONSTRAINT chk_books_status CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE INDEX idx_books_language_id ON books (language_id);
CREATE INDEX idx_books_category_id ON books (category_id);
CREATE INDEX idx_books_status ON books (status);
CREATE INDEX idx_books_deleted_at ON books (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_books_view_count ON books (view_count DESC);
CREATE INDEX idx_books_search ON books USING GIN (search);
CREATE INDEX idx_books_title_trgm ON books USING GIN (title gin_trgm_ops);

COMMENT ON COLUMN books.language_id IS 'Primary/original language metadata shown as a filter badge; verse text itself is bilingual.';
