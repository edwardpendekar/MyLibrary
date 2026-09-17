package dto

import "time"

type ImportRowResponse struct {
	RowNumber int    `json:"row_number"`
	Book      string `json:"book"`
	Chapter   int    `json:"chapter"`
	Verse     int    `json:"verse"`
	TextEN    string `json:"text_en"`
	TextID    string `json:"text_id"`
	TitleEN   string `json:"title_en"`
	TitleID   string `json:"title_id"`
}

type ImportRowErrorResponse struct {
	RowNumber int    `json:"row_number"`
	Message   string `json:"message"`
}

type ImportPreviewResponse struct {
	ImportLogID    int64                    `json:"import_log_id"`
	Status         string                   `json:"status"`
	TotalRows      int                      `json:"total_rows"`
	DistinctBooks  int                      `json:"distinct_books"`
	DistinctChapt  int                      `json:"distinct_chapters"`
	SampleRows     []ImportRowResponse      `json:"sample_rows"`
	ValidationErrs []ImportRowErrorResponse `json:"validation_errors"`
}

type CommitImportRequest struct {
	Mode string `json:"mode" validate:"required,oneof=insert upsert"`
}

type TranslateChapterRequest struct {
	BookID        int64  `json:"book_id" validate:"required"`
	ChapterNumber int    `json:"chapter_number" validate:"required,min=1"`
	TitleEN       string `json:"title_en"`
	BodyEN        string `json:"body_en" validate:"required"`
}

type ImportLogResponse struct {
	ID              int64      `json:"id"`
	Filename        string     `json:"filename"`
	Status          string     `json:"status"`
	Mode            string     `json:"mode"`
	TotalRows       int        `json:"total_rows"`
	ProcessedRows   int        `json:"processed_rows"`
	BooksCreated    int        `json:"books_created"`
	ChaptersCreated int        `json:"chapters_created"`
	SectionsCreated int        `json:"sections_created"`
	VersesInserted  int        `json:"verses_inserted"`
	VersesUpdated   int        `json:"verses_updated"`
	VersesSkipped   int        `json:"verses_skipped"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
