package dto

import "time"

type CreateBookRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Author      string  `json:"author" validate:"max=255"`
	Description string  `json:"description"`
	LanguageID  *int64  `json:"language_id"`
	CategoryID  *int64  `json:"category_id"`
	Year        *int16  `json:"year"`
	ISBN        *string `json:"isbn" validate:"omitempty,max=32"`
	Slug        string  `json:"slug" validate:"omitempty,max=255"`
}

type UpdateBookRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Author      string  `json:"author" validate:"max=255"`
	Description string  `json:"description"`
	LanguageID  *int64  `json:"language_id"`
	CategoryID  *int64  `json:"category_id"`
	Year        *int16  `json:"year"`
	ISBN        *string `json:"isbn" validate:"omitempty,max=32"`
	Slug        string  `json:"slug" validate:"omitempty,max=255"`
	Status      string  `json:"status" validate:"omitempty,oneof=draft published archived"`
}

type LanguageResponse struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
}

type CategoryResponse struct {
	ID     int64  `json:"id"`
	Slug   string `json:"slug"`
	NameEN string `json:"name_en"`
	NameID string `json:"name_id"`
}

type BookResponse struct {
	ID            int64             `json:"id"`
	Slug          string            `json:"slug"`
	Title         string            `json:"title"`
	Author        string            `json:"author"`
	Description   string            `json:"description"`
	Language      *LanguageResponse `json:"language,omitempty"`
	Category      *CategoryResponse `json:"category,omitempty"`
	Year          *int16            `json:"year"`
	ISBN          *string           `json:"isbn"`
	CoverURL      *string           `json:"cover_url"`
	Status        string            `json:"status"`
	ChaptersCount int               `json:"chapters_count"`
	VersesCount   int               `json:"verses_count"`
	ViewCount     int64             `json:"view_count"`
	HasPDF        bool              `json:"has_pdf"`
	PDFURL        *string           `json:"pdf_url,omitempty"`
	IsFavorite    bool              `json:"is_favorite,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

type ChapterResponse struct {
	ID          int64   `json:"id"`
	Number      int     `json:"number"`
	TitleEN     *string `json:"title_en,omitempty"`
	TitleID     *string `json:"title_id,omitempty"`
	VersesCount int     `json:"verses_count"`
}

type SectionResponse struct {
	ID               int64   `json:"id"`
	TitleEN          *string `json:"title_en,omitempty"`
	TitleID          *string `json:"title_id,omitempty"`
	StartVerseNumber int     `json:"start_verse_number"`
	EndVerseNumber   int     `json:"end_verse_number"`
}

type VerseResponse struct {
	ID        int64   `json:"id"`
	Number    int     `json:"number"`
	TextEN    *string `json:"text_en,omitempty"`
	TextID    *string `json:"text_id,omitempty"`
	SectionID *int64  `json:"section_id,omitempty"`
}

type ChapterContentResponse struct {
	Chapter  ChapterResponse   `json:"chapter"`
	Sections []SectionResponse `json:"sections"`
	Verses   []VerseResponse   `json:"verses"`
}
