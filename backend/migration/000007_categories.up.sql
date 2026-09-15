-- Book-level classification (e.g. "Bible", "Fiction", "History"), independent from the
-- per-verse section headings ("sections" table) that come out of the Excel title_en/title_id columns.
CREATE TABLE categories (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id   BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    slug        VARCHAR(150) NOT NULL,
    name_en     VARCHAR(150) NOT NULL,
    name_id     VARCHAR(150) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_categories_slug UNIQUE (slug)
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
