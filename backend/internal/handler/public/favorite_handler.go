package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

type FavoriteHandler struct {
	favorites *service.FavoriteService
	files     *service.FileService
}

func NewFavoriteHandler(favorites *service.FavoriteService, files *service.FileService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites, files: files}
}

// Add godoc
// @Summary      Favorite a book
// @Tags         favorites
// @Param        id path int true "book id"
// @Success      204
// @Router       /books/{id}/favorite [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.favorites.Add(c.Request.Context(), userID, bookID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// Remove godoc
// @Summary      Unfavorite a book
// @Tags         favorites
// @Param        id path int true "book id"
// @Success      204
// @Router       /books/{id}/favorite [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.favorites.Remove(c.Request.Context(), userID, bookID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// List godoc
// @Summary      List the current user's favorite books
// @Tags         favorites
// @Produce      json
// @Success      200 {object} response.Envelope{data=[]dto.BookResponse}
// @Router       /me/favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	result, err := h.favorites.List(c.Request.Context(), userID, c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := make([]dto.BookResponse, 0, len(result.Items))
	for i := range result.Items {
		b := &result.Items[i]
		resp := dto.BookResponse{
			ID: b.ID, Slug: b.Slug, Title: b.Title, Author: b.Author, Status: b.Status,
			ChaptersCount: b.ChaptersCount, VersesCount: b.VersesCount, IsFavorite: true,
		}
		if b.CoverPath != nil {
			url := h.files.URL(*b.CoverPath)
			resp.CoverURL = &url
		}
		items = append(items, resp)
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

func requireUser(c *gin.Context) (int64, bool) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, apperror.Unauthorized("authentication required"))
		return 0, false
	}
	return userID, true
}
