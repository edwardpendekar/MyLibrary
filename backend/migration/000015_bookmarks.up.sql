-- Serves both the verse reader (chapter_id/verse_id) and the PDF reader (pdf_page).
-- is_auto=true rows are the single system-maintained "last read position" per user/book
-- (used for the reader's "remember last page" feature); is_auto=false rows are user bookmarks.
CREATE TABLE bookmarks (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id    BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    chapter_id BIGINT REFERENCES chapters (id) ON DELETE CASCADE,
    verse_id   BIGINT REFERENCES verses (id) ON DELETE CASCADE,
    pdf_page   INTEGER,
    label      VARCHAR(255),
    is_auto    BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_bookmarks_target CHECK (verse_id IS NOT NULL OR pdf_page IS NOT NULL)
);

CREATE INDEX idx_bookmarks_user_book ON bookmarks (user_id, book_id);
CREATE INDEX idx_bookmarks_verse_id ON bookmarks (verse_id);

-- Only one auto-saved "last position" bookmark per user per book.
CREATE UNIQUE INDEX uq_bookmarks_auto_position
    ON bookmarks (user_id, book_id)
    WHERE is_auto = true;
