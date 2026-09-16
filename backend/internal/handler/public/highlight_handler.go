package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

type HighlightHandler struct{ highlights *service.HighlightService }

func NewHighlightHandler(highlights *service.HighlightService) *HighlightHandler {
	return &HighlightHandler{highlights: highlights}
}

// Add godoc
// @Summary      Highlight a verse
// @Tags         highlights
// @Param        id path int true "verse id"
// @Success      204
// @Router       /verses/{id}/highlight [post]
func (h *HighlightHandler) Add(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	verseID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.highlights.Add(c.Request.Context(), userID, verseID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// Remove godoc
// @Summary      Un-highlight a verse
// @Tags         highlights
// @Param        id path int true "verse id"
// @Success      204
// @Router       /verses/{id}/highlight [delete]
func (h *HighlightHandler) Remove(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	verseID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.highlights.Remove(c.Request.Context(), userID, verseID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// ListByBook godoc
// @Summary      List the verse IDs the current user has highlighted in a book
// @Tags         highlights
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=[]int64}
// @Router       /books/{id}/highlights [get]
func (h *HighlightHandler) ListByBook(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	verseIDs, err := h.highlights.ListVerseIDsByBook(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if verseIDs == nil {
		verseIDs = []int64{}
	}
	response.OK(c, verseIDs)
}
