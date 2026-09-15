CREATE TABLE favorites (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id    BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_favorites_user_book UNIQUE (user_id, book_id)
);

CREATE INDEX idx_favorites_book_id ON favorites (book_id);
