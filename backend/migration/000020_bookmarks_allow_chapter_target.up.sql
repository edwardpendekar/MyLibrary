-- The original constraint only allowed a bookmark to target a verse or a PDF
-- page. The "remember last page" auto-bookmark (is_auto=true) can also be
-- saved at chapter granularity only (verse_id/pdf_page both NULL) — see
-- VerseReader's chapter-level position tracking — so chapter_id must count
-- as a valid target too.
ALTER TABLE bookmarks DROP CONSTRAINT chk_bookmarks_target;
ALTER TABLE bookmarks ADD CONSTRAINT chk_bookmarks_target
    CHECK (verse_id IS NOT NULL OR pdf_page IS NOT NULL OR chapter_id IS NOT NULL);
