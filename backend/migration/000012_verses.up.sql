-- Hot-path table: designed for 1M+ rows, bulk COPY insert, and full-text search.
CREATE TABLE verses (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    book_id    BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    chapter_id BIGINT      NOT NULL REFERENCES chapters (id) ON DELETE CASCADE,
    section_id BIGINT REFERENCES sections (id) ON DELETE SET NULL,
    number     INTEGER     NOT NULL,
    text_en    TEXT,
    text_id    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    search     tsvector GENERATED ALWAYS AS (
                   setweight(to_tsvector('english', coalesce(text_en, '')), 'A') ||
                   setweight(to_tsvector('indonesian', coalesce(text_id, '')), 'B')
               ) STORED,
    CONSTRAINT uq_verses_chapter_number UNIQUE (chapter_id, number),
    CONSTRAINT chk_verses_number CHECK (number > 0)
);

CREATE INDEX idx_verses_book_id ON verses (book_id);
CREATE INDEX idx_verses_chapter_id ON verses (chapter_id);
CREATE INDEX idx_verses_section_id ON verses (section_id);
CREATE INDEX idx_verses_book_chapter_number ON verses (book_id, chapter_id, number);
CREATE INDEX idx_verses_search ON verses USING GIN (search);

COMMENT ON TABLE verses IS 'Leaf content unit. For very large bulk imports, indexes may be dropped and rebuilt around the COPY for throughput (see import service).';
