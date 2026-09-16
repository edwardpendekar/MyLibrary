-- Account-synced verse highlights (previously device-local only, in the
-- frontend's localStorage). Keyed on (user_id, verse_id) only — the owning
-- book is reached through verses.book_id, so listing "which verses in this
-- book has this user highlighted" is a join rather than a denormalized column.
CREATE TABLE highlights (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    verse_id   BIGINT      NOT NULL REFERENCES verses (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_highlights_user_verse UNIQUE (user_id, verse_id)
);

CREATE INDEX idx_highlights_verse_id ON highlights (verse_id);
CREATE INDEX idx_highlights_user_id ON highlights (user_id);

COMMENT ON TABLE highlights IS 'A highlight is just a (user, verse) membership row — toggled on/off, never edited.';
