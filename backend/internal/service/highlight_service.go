package service

import (
	"context"

	"bookreader/backend/internal/domain"
)

type HighlightService struct{ repo domain.HighlightRepository }

func NewHighlightService(repo domain.HighlightRepository) *HighlightService {
	return &HighlightService{repo: repo}
}

func (s *HighlightService) Add(ctx context.Context, userID, verseID int64) error {
	return s.repo.Add(ctx, userID, verseID)
}

func (s *HighlightService) Remove(ctx context.Context, userID, verseID int64) error {
	return s.repo.Remove(ctx, userID, verseID)
}

func (s *HighlightService) ListVerseIDsByBook(ctx context.Context, userID, bookID int64) ([]int64, error) {
	return s.repo.ListVerseIDsByUserAndBook(ctx, userID, bookID)
}
