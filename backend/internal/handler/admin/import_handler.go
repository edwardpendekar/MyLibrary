package admin

import (
	"time"

	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/config"
	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/pagination"
	"bookreader/backend/pkg/response"
)

const maxImportFileSize = 500 << 20 // 500MB, comfortably covers a 1M+ row spreadsheet

type ImportHandler struct {
	imports   *service.ImportService
	files     *service.FileService
	books     domain.BookRepository
	translate config.Translate
}

func NewImportHandler(imports *service.ImportService, files *service.FileService, books domain.BookRepository, translateCfg config.Translate) *ImportHandler {
	return &ImportHandler{imports: imports, files: files, books: books, translate: translateCfg}
}

// Upload godoc
// @Summary      Upload an Excel/CSV file and get a validation preview
// @Tags         admin-import
// @Accept       multipart/form-data
// @Param        file formData file true "xlsx, xls, or csv (Book, Chapter, Verse, text_en, text_id, title_en, title_id)"
// @Success      200 {object} response.Envelope{data=dto.ImportPreviewResponse}
// @Router       /admin/import/upload [post]
func (h *ImportHandler) Upload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, apperror.Validation("file is required", nil))
		return
	}
	if header.Size > maxImportFileSize {
		response.Fail(c, apperror.Validation("file must be 500MB or smaller", nil))
		return
	}

	userID, _ := middleware.UserID(c)
	stream, err := header.Open()
	if err != nil {
		response.Fail(c, apperror.Internal("failed to read uploaded file", err))
		return
	}
	defer stream.Close()

	uploaded, err := h.files.Upload(c.Request.Context(), service.UploadInput{
		Folder: "imports", Filename: header.Filename, ContentType: header.Header.Get("Content-Type"),
		Size: header.Size, Reader: stream, UploadedBy: userID,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}

	log, preview, err := h.imports.Upload(c.Request.Context(), uploaded, userID, header.Filename)
	if err != nil {
		response.Fail(c, err)
		return
	}

	resp := dto.ImportPreviewResponse{
		ImportLogID: log.ID, Status: log.Status,
		// Never nil: Go serializes a nil slice as JSON null, not [], which
		// crashes naive frontend code doing `.length` on an "always an array" field.
		SampleRows:     []dto.ImportRowResponse{},
		ValidationErrs: []dto.ImportRowErrorResponse{},
	}
	if preview != nil {
		resp.TotalRows = preview.TotalRows
		resp.DistinctBooks = preview.DistinctBooks
		resp.DistinctChapt = preview.DistinctChapt
		for _, r := range preview.SampleRows {
			resp.SampleRows = append(resp.SampleRows, dto.ImportRowResponse{
				RowNumber: r.RowNumber, Book: r.Book, Chapter: r.Chapter, Verse: r.Verse,
				TextEN: r.TextEN, TextID: r.TextID, TitleEN: r.TitleEN, TitleID: r.TitleID,
			})
		}
		for _, e := range preview.ValidationErrs {
			resp.ValidationErrs = append(resp.ValidationErrs, dto.ImportRowErrorResponse{RowNumber: e.RowNumber, Message: e.Message})
		}
	}
	response.OK(c, resp)
}

// Translate godoc
// @Summary      Translate one chapter's pasted English text into an importable CSV
// @Tags         admin-import
// @Accept       json
// @Param        request body dto.TranslateChapterRequest true "book, chapter number, and English title/body"
// @Success      200 {object} response.Envelope{data=dto.ImportLogResponse}
// @Router       /admin/import/translate [post]
func (h *ImportHandler) Translate(c *gin.Context) {
	var req dto.TranslateChapterRequest
	if !httpx.BindJSON(c, &req) {
		return
	}

	book, err := h.books.FindByID(c.Request.Context(), req.BookID)
	if err != nil || book == nil {
		response.Fail(c, apperror.NotFound("book not found"))
		return
	}

	userID, _ := middleware.UserID(c)
	log, err := h.imports.Translate(c.Request.Context(), service.TranslateConfig{
		PythonBin: h.translate.PythonBin, ScriptPath: h.translate.ScriptPath,
		GeminiAPIKey: h.translate.GeminiAPIKey, GeminiModel: h.translate.GeminiModel,
		CommandTimeout: h.translate.CommandTimeout,
	}, userID, book.Title, req.ChapterNumber, req.TitleEN, req.BodyEN)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToImportLogResponse(log))
}

// Preview godoc
// @Summary      Re-fetch the validation preview for an import job once it's ready
// @Tags         admin-import
// @Param        id path int true "import log id"
// @Success      200 {object} response.Envelope{data=dto.ImportPreviewResponse}
// @Router       /admin/import/{id}/preview [get]
func (h *ImportHandler) Preview(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	log, preview, err := h.imports.Preview(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	resp := dto.ImportPreviewResponse{
		ImportLogID: log.ID, Status: log.Status,
		SampleRows:     []dto.ImportRowResponse{},
		ValidationErrs: []dto.ImportRowErrorResponse{},
	}
	if preview != nil {
		resp.TotalRows = preview.TotalRows
		resp.DistinctBooks = preview.DistinctBooks
		resp.DistinctChapt = preview.DistinctChapt
		for _, r := range preview.SampleRows {
			resp.SampleRows = append(resp.SampleRows, dto.ImportRowResponse{
				RowNumber: r.RowNumber, Book: r.Book, Chapter: r.Chapter, Verse: r.Verse,
				TextEN: r.TextEN, TextID: r.TextID, TitleEN: r.TitleEN, TitleID: r.TitleID,
			})
		}
		for _, e := range preview.ValidationErrs {
			resp.ValidationErrs = append(resp.ValidationErrs, dto.ImportRowErrorResponse{RowNumber: e.RowNumber, Message: e.Message})
		}
	}
	response.OK(c, resp)
}

// Commit godoc
// @Summary      Start the actual import for a previously uploaded/validated file
// @Tags         admin-import
// @Accept       json
// @Param        id path int true "import log id"
// @Success      202 {object} response.Envelope{data=dto.ImportLogResponse}
// @Router       /admin/import/{id}/commit [post]
func (h *ImportHandler) Commit(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	var req dto.CommitImportRequest
	if !httpx.BindJSON(c, &req) {
		return
	}

	log, err := h.imports.GetStatus(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.imports.Commit(id, req.Mode); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToImportLogResponse(log))
}

// Status godoc
// @Summary      Poll an import job's progress
// @Tags         admin-import
// @Param        id path int true "import log id"
// @Success      200 {object} response.Envelope{data=dto.ImportLogResponse}
// @Router       /admin/import/{id} [get]
func (h *ImportHandler) Status(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	log, err := h.imports.GetStatus(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToImportLogResponse(log))
}

// Stream godoc
// @Summary      Server-Sent Events progress stream for an import job
// @Tags         admin-import
// @Param        id path int true "import log id"
// @Produce      text/event-stream
// @Router       /admin/import/{id}/stream [get]
func (h *ImportHandler) Stream(c *gin.Context) {
	id, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			log, err := h.imports.GetStatus(c.Request.Context(), id)
			if err != nil {
				c.SSEvent("error", err.Error())
				c.Writer.Flush()
				return
			}
			c.SSEvent("progress", dto.ToImportLogResponse(log))
			c.Writer.Flush()
			if isTerminalStatus(log.Status) {
				return
			}
		}
	}
}

// List godoc
// @Summary      Import job history
// @Tags         admin-import
// @Success      200 {object} response.Envelope{data=[]dto.ImportLogResponse}
// @Router       /admin/import [get]
func (h *ImportHandler) List(c *gin.Context) {
	result, err := h.imports.List(c.Request.Context(), c.Query("cursor"), httpx.QueryInt(c, "limit", pagination.DefaultLimit))
	if err != nil {
		response.Fail(c, err)
		return
	}
	items := make([]dto.ImportLogResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, dto.ToImportLogResponse(&result.Items[i]))
	}
	response.OKWithMeta(c, items, pagination.Page{NextCursor: result.NextCursor, HasMore: result.HasMore, Limit: len(items)})
}

// isTerminalStatus tells the SSE loop when to stop polling. "ready" is
// included alongside the normal terminal statuses because it also ends a
// stream started by Translate (translating -> ready), even though the plain
// CSV upload flow never opens a stream while status is "ready" (its preview
// is returned synchronously by Upload instead).
func isTerminalStatus(status string) bool {
	switch status {
	case "ready", "completed", "failed", "rolled_back":
		return true
	default:
		return false
	}
}
