CREATE TABLE notes (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id    BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    verse_id   BIGINT REFERENCES verses (id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notes_user_book ON notes (user_id, book_id);
CREATE INDEX idx_notes_verse_id ON notes (verse_id);
