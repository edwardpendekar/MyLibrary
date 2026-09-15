package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/storage"
)

// FileService is the single place that turns an uploaded stream into a
// storage.Provider object plus a files-table row, used by both cover and PDF
// uploads (and, later, the Excel import upload).
type FileService struct {
	files    domain.FileRepository
	pdfs     domain.PDFRepository
	provider storage.Provider
}

func NewFileService(files domain.FileRepository, pdfs domain.PDFRepository, provider storage.Provider) *FileService {
	return &FileService{files: files, pdfs: pdfs, provider: provider}
}

type UploadInput struct {
	Folder      string
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
	UploadedBy  int64
}

func (s *FileService) Upload(ctx context.Context, in UploadInput) (*domain.File, error) {
	hasher := sha256.New()
	tee := io.TeeReader(in.Reader, hasher)

	meta, err := s.provider.Save(ctx, in.Folder, in.Filename, tee, in.Size, in.ContentType)
	if err != nil {
		return nil, apperror.Internal("failed to store file", err)
	}

	file := &domain.File{
		UploadedBy: &in.UploadedBy, OriginalName: in.Filename, StoredPath: meta.Key,
		Provider: providerName(s.provider), MimeType: in.ContentType, SizeBytes: meta.Size,
		ChecksumSHA256: hex.EncodeToString(hasher.Sum(nil)),
	}
	if err := s.files.Create(ctx, file); err != nil {
		return nil, apperror.Internal("failed to record file", err)
	}
	return file, nil
}

func (s *FileService) AttachPDF(ctx context.Context, bookID, fileID int64, pageCount *int) error {
	return s.pdfs.Upsert(ctx, &domain.PDF{BookID: bookID, FileID: fileID, PageCount: pageCount})
}

func (s *FileService) URL(key string) string {
	return s.provider.URL(key)
}

func providerName(p storage.Provider) string {
	switch p.(type) {
	case *storage.S3:
		return "s3"
	default:
		return "local"
	}
}
