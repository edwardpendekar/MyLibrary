package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

type SearchHandler struct{ search *service.SearchService }

func NewSearchHandler(search *service.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

// SearchVerses godoc
// @Summary      Full-text search across verse text (EN + ID) and titles
// @Tags         search
// @Produce      json
// @Param        q query string true "search query"
// @Param        book_id query int false "restrict to one book"
// @Param        cursor query string false "pagination cursor"
// @Param        limit query int false "page size"
// @Success      200 {object} response.Envelope{data=[]dto.SearchVerseHitResponse}
// @Router       /search/verses [get]
func (h *SearchHandler) SearchVerses(c *gin.Context) {
	q := c.Query("q")
	bookID := httpx.QueryInt64Ptr(c, "book_id")
	limit := httpx.QueryInt(c, "limit", pagination.DefaultLimit)

	result, err := h.search.SearchVerses(c.Request.Context(), q, c.Query("lang"), bookID, c.Query("cursor"), limit)
	if err != nil {
		response.Fail(c, err)
		return
	}

	items := make([]dto.SearchVerseHitResponse, 0, len(result.Items))
	for _, hit := range result.Items {
		items = append(items, dto.SearchVerseHitResponse{
			BookID: hit.BookID, BookTitle: hit.BookTitle, BookSlug: hit.BookSlug,
			ChapterID: hit.ChapterID, ChapterNum: hit.ChapterNum, VerseID: hit.VerseID,
			VerseNumber: hit.VerseNumber, Snippet: hit.Snippet, Rank: hit.Rank,
		})
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

// SearchBooks godoc
// @Summary      Full-text search across book title/author/description
// @Tags         search
// @Produce      json
// @Param        q query string true "search query"
// @Success      200 {object} response.Envelope{data=[]dto.BookResponse}
// @Router       /search/books [get]
func (h *SearchHandler) SearchBooks(c *gin.Context) {
	limit := httpx.QueryInt(c, "limit", pagination.DefaultLimit)
	result, err := h.search.SearchBooks(c.Request.Context(), c.Query("q"), c.Query("cursor"), limit)
	if err != nil {
		response.Fail(c, err)
		return
	}

	items := make([]dto.BookResponse, 0, len(result.Items))
	for i := range result.Items {
		b := &result.Items[i]
		items = append(items, dto.BookResponse{
			ID: b.ID, Slug: b.Slug, Title: b.Title, Author: b.Author, Status: b.Status,
			ChaptersCount: b.ChaptersCount, VersesCount: b.VersesCount,
		})
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}
