package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

type BookHandler struct {
	books     *service.BookService
	favorites *service.FavoriteService
	files     *service.FileService
	pdfs      domain.PDFRepository
}

func NewBookHandler(books *service.BookService, favorites *service.FavoriteService, files *service.FileService, pdfs domain.PDFRepository) *BookHandler {
	return &BookHandler{books: books, favorites: favorites, files: files, pdfs: pdfs}
}

// List godoc
// @Summary      List published books (Home grid)
// @Tags         books
// @Produce      json
// @Param        cursor query string false "pagination cursor"
// @Param        limit query int false "page size"
// @Param        category_id query int false "filter by category"
// @Param        language_id query int false "filter by language"
// @Param        q query string false "title/author search"
// @Success      200 {object} response.Envelope{data=[]dto.BookResponse}
// @Router       /books [get]
func (h *BookHandler) List(c *gin.Context) {
	published := domain.BookStatusPublished
	filter := domain.BookFilter{
		Status:     &published,
		CategoryID: httpx.QueryInt64Ptr(c, "category_id"),
		LanguageID: httpx.QueryInt64Ptr(c, "language_id"),
		Query:      c.Query("q"),
	}

	result, err := h.books.List(c.Request.Context(), filter, c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}

	items := h.toResponses(c, result.Items)
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

// Popular godoc
// @Summary      Most-viewed published books, for the Home page "Popular" rail
// @Tags         books
// @Produce      json
// @Success      200 {object} response.Envelope{data=[]dto.BookResponse}
// @Router       /books/popular [get]
func (h *BookHandler) Popular(c *gin.Context) {
	books, err := h.books.ListPopular(c.Request.Context(), httpx.QueryInt(c, "limit", 10))
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := h.toResponses(c, books)
	response.OK(c, items)
}

// Detail godoc
// @Summary      Book detail page
// @Tags         books
// @Produce      json
// @Param        slug path string true "book slug"
// @Success      200 {object} response.Envelope{data=dto.BookResponse}
// @Failure      404 {object} response.Envelope
// @Router       /books/{slug} [get]
func (h *BookHandler) Detail(c *gin.Context) {
	book, err := h.books.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	pdf, _ := h.pdfs.FindByBookID(c.Request.Context(), book.ID)
	resp := h.toResponse(c, book, pdf)
	response.OK(c, resp)
}

// Chapters godoc
// @Summary      List a book's chapters (reader's left panel)
// @Tags         books
// @Produce      json
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=[]dto.ChapterResponse}
// @Router       /books/{id}/chapters [get]
func (h *BookHandler) Chapters(c *gin.Context) {
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	chapters, err := h.books.ListChapters(c.Request.Context(), bookID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := make([]dto.ChapterResponse, 0, len(chapters))
	for i := range chapters {
		items = append(items, dto.ToChapterResponse(&chapters[i]))
	}
	response.OK(c, items)
}

// ChapterContent godoc
// @Summary      Chapter content: sections + verses (reader's center panel)
// @Tags         books
// @Produce      json
// @Param        id path int true "book id"
// @Param        number path int true "chapter number"
// @Success      200 {object} response.Envelope{data=dto.ChapterContentResponse}
// @Router       /books/{id}/chapters/{number} [get]
func (h *BookHandler) ChapterContent(c *gin.Context) {
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	number, ok := httpx.ParamInt64(c, "number")
	if !ok {
		return
	}

	content, err := h.books.GetChapterContent(c.Request.Context(), bookID, int(number))
	if err != nil {
		response.Fail(c, err)
		return
	}

	sections := make([]dto.SectionResponse, 0, len(content.Sections))
	for i := range content.Sections {
		sections = append(sections, dto.ToSectionResponse(&content.Sections[i]))
	}
	verses := make([]dto.VerseResponse, 0, len(content.Verses))
	for i := range content.Verses {
		verses = append(verses, dto.ToVerseResponse(&content.Verses[i]))
	}

	response.OK(c, dto.ChapterContentResponse{
		Chapter:  dto.ToChapterResponse(&content.Chapter),
		Sections: sections,
		Verses:   verses,
	})
}

// toResponses batch-loads PDFs for the whole page of books (one query) instead
// of querying per book, then delegates to toResponse for the per-book shaping.
func (h *BookHandler) toResponses(c *gin.Context, books []domain.Book) []dto.BookResponse {
	ids := make([]int64, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}
	pdfByBook, _ := h.pdfs.FindByBookIDs(c.Request.Context(), ids)

	items := make([]dto.BookResponse, 0, len(books))
	for i := range books {
		var pdf *domain.PDF
		if p, ok := pdfByBook[books[i].ID]; ok {
			pdf = &p
		}
		items = append(items, h.toResponse(c, &books[i], pdf))
	}
	return items
}

func (h *BookHandler) toResponse(c *gin.Context, b *domain.Book, pdf *domain.PDF) dto.BookResponse {
	resp := dto.BookResponse{
		ID: b.ID, Slug: b.Slug, Title: b.Title, Author: b.Author, Description: b.Description,
		Language: dto.ToLanguageResponse(b.Language), Category: dto.ToCategoryResponse(b.Category),
		Year: b.Year, ISBN: b.ISBN, Status: b.Status,
		ChaptersCount: b.ChaptersCount, VersesCount: b.VersesCount, ViewCount: b.ViewCount,
		CreatedAt: b.CreatedAt,
	}
	if b.CoverPath != nil && *b.CoverPath != "" {
		url := h.files.URL(*b.CoverPath)
		resp.CoverURL = &url
	}
	if pdf != nil {
		resp.HasPDF = true
		url := h.files.URL(pdf.File.StoredPath)
		resp.PDFURL = &url
	}
	if userID, ok := middleware.UserID(c); ok {
		isFav, _ := h.favorites.IsFavorite(c.Request.Context(), userID, b.ID)
		resp.IsFavorite = isFav
	}
	return resp
}
