package dto

import "time"

type CreateBookmarkRequest struct {
	BookID    int64   `json:"book_id" validate:"required"`
	ChapterID *int64  `json:"chapter_id"`
	VerseID   *int64  `json:"verse_id"`
	PDFPage   *int    `json:"pdf_page"`
	Label     *string `json:"label" validate:"omitempty,max=255"`
}

type SaveLastPositionRequest struct {
	BookID    int64  `json:"book_id" validate:"required"`
	ChapterID *int64 `json:"chapter_id"`
	VerseID   *int64 `json:"verse_id"`
	PDFPage   *int   `json:"pdf_page"`
}

type BookmarkResponse struct {
	ID        int64     `json:"id"`
	BookID    int64     `json:"book_id"`
	ChapterID *int64    `json:"chapter_id,omitempty"`
	VerseID   *int64    `json:"verse_id,omitempty"`
	PDFPage   *int      `json:"pdf_page,omitempty"`
	Label     *string   `json:"label,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateNoteRequest struct {
	BookID  int64  `json:"book_id" validate:"required"`
	VerseID *int64 `json:"verse_id"`
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type UpdateNoteRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type NoteResponse struct {
	ID        int64     `json:"id"`
	BookID    int64     `json:"book_id"`
	VerseID   *int64    `json:"verse_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
