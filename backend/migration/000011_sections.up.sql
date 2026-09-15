-- Normalizes the repeated title_en/title_id heading that the Excel import carries on every
-- verse row of the same passage (e.g. "Creation" / "Penciptaan" repeated for Genesis 1:1-1:3),
-- instead of duplicating that text across a million verse rows.
CREATE TABLE sections (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    book_id            BIGINT       NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    chapter_id         BIGINT       NOT NULL REFERENCES chapters (id) ON DELETE CASCADE,
    title_en           VARCHAR(255),
    title_id           VARCHAR(255),
    order_index        INTEGER      NOT NULL,
    start_verse_number INTEGER      NOT NULL,
    end_verse_number   INTEGER      NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_sections_chapter_order UNIQUE (chapter_id, order_index)
);

CREATE INDEX idx_sections_book_id ON sections (book_id);
CREATE INDEX idx_sections_chapter_id ON sections (chapter_id);
