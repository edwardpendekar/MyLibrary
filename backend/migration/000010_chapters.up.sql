CREATE TABLE chapters (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    book_id      BIGINT       NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    number       INTEGER      NOT NULL,
    title_en     VARCHAR(255),
    title_id     VARCHAR(255),
    verses_count INTEGER      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_chapters_book_number UNIQUE (book_id, number),
    CONSTRAINT chk_chapters_number CHECK (number > 0)
);

CREATE INDEX idx_chapters_book_id ON chapters (book_id);
