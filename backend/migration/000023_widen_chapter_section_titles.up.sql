-- Chapter/section "titles" in real imported content are often full
-- descriptive sentences (patristic/theological texts routinely have
-- section headings well past 255 characters), not short headings. The
-- VARCHAR(255) cap made a large real-world import abort entirely on the
-- first over-length title it hit. Verse text (verses.text_en/text_id) was
-- already unbounded TEXT for the same reason; titles get the same treatment.
ALTER TABLE chapters ALTER COLUMN title_en TYPE TEXT;
ALTER TABLE chapters ALTER COLUMN title_id TYPE TEXT;
ALTER TABLE sections ALTER COLUMN title_en TYPE TEXT;
ALTER TABLE sections ALTER COLUMN title_id TYPE TEXT;
