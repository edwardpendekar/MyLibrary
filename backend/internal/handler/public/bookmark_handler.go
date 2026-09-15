package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

type BookmarkHandler struct{ bookmarks *service.BookmarkService }

func NewBookmarkHandler(bookmarks *service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{bookmarks: bookmarks}
}

// Create godoc
// @Summary      Create a bookmark (verse or PDF page)
// @Tags         bookmarks
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateBookmarkRequest true "bookmark"
// @Success      201 {object} response.Envelope{data=dto.BookmarkResponse}
// @Router       /bookmarks [post]
func (h *BookmarkHandler) Create(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req dto.CreateBookmarkRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	bookmark := &domain.Bookmark{
		UserID: userID, BookID: req.BookID, ChapterID: req.ChapterID,
		VerseID: req.VerseID, PDFPage: req.PDFPage, Label: req.Label,
	}
	if err := h.bookmarks.Create(c.Request.Context(), bookmark); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, dto.ToBookmarkResponse(bookmark))
}

// Delete godoc
// @Summary      Delete a bookmark
// @Tags         bookmarks
// @Param        id path int true "bookmark id"
// @Success      204
// @Router       /bookmarks/{id} [delete]
func (h *BookmarkHandler) Delete(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.bookmarks.Delete(c.Request.Context(), id, userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// ListByBook godoc
// @Summary      List the user's bookmarks for a book
// @Tags         bookmarks
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=[]dto.BookmarkResponse}
// @Router       /books/{id}/bookmarks [get]
func (h *BookmarkHandler) ListByBook(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	items, err := h.bookmarks.ListByBook(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	out := make([]dto.BookmarkResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.ToBookmarkResponse(&items[i]))
	}
	response.OK(c, out)
}

// SaveLastPosition godoc
// @Summary      Save/update the "remember last page" auto-bookmark
// @Tags         bookmarks
// @Accept       json
// @Success      204
// @Router       /bookmarks/last-position [put]
func (h *BookmarkHandler) SaveLastPosition(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req dto.SaveLastPositionRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	bookmark := &domain.Bookmark{
		UserID: userID, BookID: req.BookID, ChapterID: req.ChapterID, VerseID: req.VerseID, PDFPage: req.PDFPage,
	}
	if err := h.bookmarks.SaveLastPosition(c.Request.Context(), bookmark); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// LastPosition godoc
// @Summary      Get the "remember last page" position for a book
// @Tags         bookmarks
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=dto.BookmarkResponse}
// @Router       /books/{id}/last-position [get]
func (h *BookmarkHandler) LastPosition(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	pos, err := h.bookmarks.LastPosition(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if pos == nil {
		response.OK(c, nil)
		return
	}
	response.OK(c, dto.ToBookmarkResponse(pos))
}
