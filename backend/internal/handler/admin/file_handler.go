package admin

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

const (
	maxCoverSize = 5 << 20   // 5MB
	maxPDFSize   = 200 << 20 // 200MB
)

type FileHandler struct {
	files domain.FileRepository
	books domain.BookRepository
	fs    *service.FileService
}

func NewFileHandler(files domain.FileRepository, books domain.BookRepository, fs *service.FileService) *FileHandler {
	return &FileHandler{files: files, books: books, fs: fs}
}

// UploadCover godoc
// @Summary      Upload a book cover image
// @Tags         admin-files
// @Accept       multipart/form-data
// @Param        id path int true "book id"
// @Param        file formData file true "cover image (jpg/png/webp, max 5MB)"
// @Success      200 {object} response.Envelope
// @Router       /admin/books/{id}/cover [post]
func (h *FileHandler) UploadCover(c *gin.Context) {
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	book, err := h.books.FindByID(c.Request.Context(), bookID)
	if err != nil || book == nil {
		response.Fail(c, apperror.NotFound("book not found"))
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, apperror.Validation("file is required", nil))
		return
	}
	if header.Size > maxCoverSize {
		response.Fail(c, apperror.Validation("cover image must be 5MB or smaller", nil))
		return
	}
	if !isAllowedImage(header) {
		response.Fail(c, apperror.Validation("cover must be a JPEG, PNG, or WebP image", nil))
		return
	}

	userID, _ := middleware.UserID(c)
	stream, err := header.Open()
	if err != nil {
		response.Fail(c, apperror.Internal("failed to read uploaded file", err))
		return
	}
	defer stream.Close()

	file, err := h.fs.Upload(c.Request.Context(), service.UploadInput{
		Folder: "covers", Filename: header.Filename, ContentType: header.Header.Get("Content-Type"),
		Size: header.Size, Reader: stream, UploadedBy: userID,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}

	book.CoverPath = &file.StoredPath
	book.CoverFileID = &file.ID
	if err := h.books.Update(c.Request.Context(), book); err != nil {
		response.Fail(c, apperror.Internal("failed to attach cover to book", err))
		return
	}
	response.OK(c, gin.H{"cover_url": h.fs.URL(file.StoredPath)})
}

// UploadPDF godoc
// @Summary      Upload a book's PDF ebook
// @Tags         admin-files
// @Accept       multipart/form-data
// @Param        id path int true "book id"
// @Param        file formData file true "PDF file (max 200MB)"
// @Success      200 {object} response.Envelope
// @Router       /admin/books/{id}/pdf [post]
func (h *FileHandler) UploadPDF(c *gin.Context) {
	bookID, ok := httpx.ParamInt64(c, "id")
	if !ok {
		return
	}
	book, err := h.books.FindByID(c.Request.Context(), bookID)
	if err != nil || book == nil {
		response.Fail(c, apperror.NotFound("book not found"))
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, apperror.Validation("file is required", nil))
		return
	}
	if header.Size > maxPDFSize {
		response.Fail(c, apperror.Validation("PDF must be 200MB or smaller", nil))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		response.Fail(c, apperror.Validation("file must be a PDF", nil))
		return
	}

	userID, _ := middleware.UserID(c)
	stream, err := header.Open()
	if err != nil {
		response.Fail(c, apperror.Internal("failed to read uploaded file", err))
		return
	}
	defer stream.Close()

	file, err := h.fs.Upload(c.Request.Context(), service.UploadInput{
		Folder: "pdfs", Filename: header.Filename, ContentType: contentType,
		Size: header.Size, Reader: stream, UploadedBy: userID,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.fs.AttachPDF(c.Request.Context(), bookID, file.ID, nil); err != nil {
		response.Fail(c, apperror.Internal("failed to attach PDF to book", err))
		return
	}
	response.OK(c, gin.H{"pdf_url": h.fs.URL(file.StoredPath)})
}

func isAllowedImage(header *multipart.FileHeader) bool {
	switch header.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
