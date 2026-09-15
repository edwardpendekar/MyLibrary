ALTER TABLE bookmarks DROP CONSTRAINT chk_bookmarks_target;
ALTER TABLE bookmarks ADD CONSTRAINT chk_bookmarks_target
    CHECK (verse_id IS NOT NULL OR pdf_page IS NOT NULL);
