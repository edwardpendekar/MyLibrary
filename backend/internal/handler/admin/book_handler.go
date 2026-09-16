// Package admin exposes role-gated management endpoints (RequireRole(admin, editor)
// applied in the router) for books, reference data, users, imports, and files.
package admin

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
	books *service.BookService
	files *service.FileService
	pdfs  domain.PDFRepository
}

func NewBookHandler(books *service.BookService, files *service.FileService, pdfs domain.PDFRepository) *BookHandler {
	return &BookHandler{books: books, files: files, pdfs: pdfs}
}

// List godoc
// @Summary      List all books (any status) for the admin table
// @Tags         admin-books
// @Produce      json
// @Success      200 {object} response.Envelope{data=[]dto.BookResponse}
// @Router       /admin/books [get]
func (h *BookHandler) List(c *gin.Context) {
	filter := domain.BookFilter{
		CategoryID: httpx.QueryInt64Ptr(c, "category_id"),
		LanguageID: httpx.QueryInt64Ptr(c, "language_id"),
		Query:      c.Query("q"),
	}
	result, err := h.books.List(c.Request.Context(), filter, c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}
	ids := make([]int64, len(result.Items))
	for i, b := range result.Items {
		ids[i] = b.ID
	}
	pdfByBook, _ := h.pdfs.FindByBookIDs(c.Request.Context(), ids)

	items := make([]dto.BookResponse, 0, len(result.Items))
	for i := range result.Items {
		var pdf *domain.PDF
		if p, ok := pdfByBook[result.Items[i].ID]; ok {
			pdf = &p
		}
		items = append(items, toAdminBookResponse(h.files, &result.Items[i], pdf))
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

// Create godoc
// @Summary      Create a book
// @Tags         admin-books
// @Accept       json
// @Success      201 {object} response.Envelope{data=dto.BookResponse}
// @Router       /admin/books [post]
func (h *BookHandler) Create(c *gin.Context) {
	var req dto.CreateBookRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	userID, _ := middleware.UserID(c)
	book := &domain.Book{
		Title: req.Title, Author: req.Author, Description: req.Description,
		LanguageID: req.LanguageID, CategoryID: req.CategoryID, Year: req.Year, ISBN: req.ISBN, Slug: req.Slug,
	}
	if err := h.books.Create(c.Request.Context(), book, userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, toAdminBookResponse(h.files, book, nil))
}

// Update godoc
// @Summary      Update a book
// @Tags         admin-books
// @Accept       json
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=dto.BookResponse}
// @Router       /admin/books/{id} [put]
func (h *BookHandler) Update(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateBookRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	// Loaded first (rather than building a bare struct from the request) so
	// fields this form doesn't edit — cover/PDF references, view count —
	// survive the update instead of being zeroed out by Update's blanket
	// column overwrite.
	book, err := h.books.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	book.Title, book.Author, book.Description = req.Title, req.Author, req.Description
	book.LanguageID, book.CategoryID = req.LanguageID, req.CategoryID
	book.Year, book.ISBN, book.Slug, book.Status = req.Year, req.ISBN, req.Slug, req.Status
	if err := h.books.Update(c.Request.Context(), book); err != nil {
		response.Fail(c, err)
		return
	}
	pdf, _ := h.pdfs.FindByBookID(c.Request.Context(), book.ID)
	response.OK(c, toAdminBookResponse(h.files, book, pdf))
}

// Delete godoc
// @Summary      Soft-delete a book
// @Tags         admin-books
// @Param        id path int true "book id"
// @Success      204
// @Router       /admin/books/{id} [delete]
func (h *BookHandler) Delete(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.books.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// Detail godoc
// @Summary      Get a book by ID (admin edit form)
// @Tags         admin-books
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=dto.BookResponse}
// @Router       /admin/books/{id} [get]
func (h *BookHandler) Detail(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	book, err := h.books.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	pdf, _ := h.pdfs.FindByBookID(c.Request.Context(), book.ID)
	response.OK(c, toAdminBookResponse(h.files, book, pdf))
}

func toAdminBookResponse(files *service.FileService, b *domain.Book, pdf *domain.PDF) dto.BookResponse {
	resp := dto.BookResponse{
		ID: b.ID, Slug: b.Slug, Title: b.Title, Author: b.Author, Description: b.Description,
		Language: dto.ToLanguageResponse(b.Language), Category: dto.ToCategoryResponse(b.Category),
		Year: b.Year, ISBN: b.ISBN, Status: b.Status,
		ChaptersCount: b.ChaptersCount, VersesCount: b.VersesCount, ViewCount: b.ViewCount,
		CreatedAt: b.CreatedAt,
	}
	if b.CoverPath != nil && *b.CoverPath != "" {
		url := files.URL(*b.CoverPath)
		resp.CoverURL = &url
	}
	if pdf != nil {
		resp.HasPDF = true
		url := files.URL(pdf.File.StoredPath)
		resp.PDFURL = &url
	}
	return resp
}
