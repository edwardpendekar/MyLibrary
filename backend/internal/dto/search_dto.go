package dto

type SearchVerseHitResponse struct {
	BookID      int64   `json:"book_id"`
	BookTitle   string  `json:"book_title"`
	BookSlug    string  `json:"book_slug"`
	ChapterID   *int64  `json:"chapter_id,omitempty"`
	ChapterNum  *int    `json:"chapter_number,omitempty"`
	VerseID     *int64  `json:"verse_id,omitempty"`
	VerseNumber *int    `json:"verse_number,omitempty"`
	Snippet     string  `json:"snippet"`
	Rank        float64 `json:"rank"`
}
