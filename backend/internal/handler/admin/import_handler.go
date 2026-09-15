package admin

import (
	"time"

	"github.com/gin-gonic/gin"

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
	imports *service.ImportService
	files   *service.FileService
}

func NewImportHandler(imports *service.ImportService, files *service.FileService) *ImportHandler {
	return &ImportHandler{imports: imports, files: files}
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

func isTerminalStatus(status string) bool {
	switch status {
	case "completed", "failed", "rolled_back":
		return true
	default:
		return false
	}
}
