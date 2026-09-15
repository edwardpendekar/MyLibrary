package public

import (
	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

type NoteHandler struct{ notes *service.NoteService }

func NewNoteHandler(notes *service.NoteService) *NoteHandler { return &NoteHandler{notes: notes} }

// Create godoc
// @Summary      Create a note on a book or verse
// @Tags         notes
// @Accept       json
// @Success      201 {object} response.Envelope{data=dto.NoteResponse}
// @Router       /notes [post]
func (h *NoteHandler) Create(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req dto.CreateNoteRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	note := &domain.Note{UserID: userID, BookID: req.BookID, VerseID: req.VerseID, Content: req.Content}
	if err := h.notes.Create(c.Request.Context(), note); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, dto.ToNoteResponse(note))
}

// Update godoc
// @Summary      Edit a note
// @Tags         notes
// @Accept       json
// @Param        id path int true "note id"
// @Success      200 {object} response.Envelope{data=dto.NoteResponse}
// @Router       /notes/{id} [put]
func (h *NoteHandler) Update(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateNoteRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	note := &domain.Note{ID: id, UserID: userID, Content: req.Content}
	if err := h.notes.Update(c.Request.Context(), note); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToNoteResponse(note))
}

// Delete godoc
// @Summary      Delete a note
// @Tags         notes
// @Param        id path int true "note id"
// @Success      204
// @Router       /notes/{id} [delete]
func (h *NoteHandler) Delete(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	if err := h.notes.Delete(c.Request.Context(), id, userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// ListByBook godoc
// @Summary      List the user's notes for a book
// @Tags         notes
// @Param        id path int true "book id"
// @Success      200 {object} response.Envelope{data=[]dto.NoteResponse}
// @Router       /books/{id}/notes [get]
func (h *NoteHandler) ListByBook(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	items, err := h.notes.ListByBook(c.Request.Context(), userID, bookID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	out := make([]dto.NoteResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.ToNoteResponse(&items[i]))
	}
	response.OK(c, out)
}
