package service

import (
	"context"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
)

type BookmarkService struct{ repo domain.BookmarkRepository }

func NewBookmarkService(repo domain.BookmarkRepository) *BookmarkService {
	return &BookmarkService{repo: repo}
}

func (s *BookmarkService) Create(ctx context.Context, b *domain.Bookmark) error {
	if b.VerseID == nil && b.PDFPage == nil {
		return apperror.Validation("bookmark must target a verse or a PDF page", nil)
	}
	return s.repo.Create(ctx, b)
}

func (s *BookmarkService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *BookmarkService) ListByBook(ctx context.Context, userID, bookID int64) ([]domain.Bookmark, error) {
	return s.repo.ListByUserAndBook(ctx, userID, bookID)
}

// SaveLastPosition powers "remember last page/verse": called by the frontend on a
// debounce as the reader scrolls, so the next visit resumes where the user left off.
func (s *BookmarkService) SaveLastPosition(ctx context.Context, b *domain.Bookmark) error {
	return s.repo.UpsertLastPosition(ctx, b)
}

func (s *BookmarkService) LastPosition(ctx context.Context, userID, bookID int64) (*domain.Bookmark, error) {
	return s.repo.FindLastPosition(ctx, userID, bookID)
}
