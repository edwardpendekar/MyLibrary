package service

import (
	"context"

	"bookreader/backend/internal/domain"
)

type NoteService struct{ repo domain.NoteRepository }

func NewNoteService(repo domain.NoteRepository) *NoteService { return &NoteService{repo: repo} }

func (s *NoteService) Create(ctx context.Context, n *domain.Note) error {
	return s.repo.Create(ctx, n)
}

func (s *NoteService) Update(ctx context.Context, n *domain.Note) error {
	return s.repo.Update(ctx, n)
}

func (s *NoteService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *NoteService) ListByBook(ctx context.Context, userID, bookID int64) ([]domain.Note, error) {
	return s.repo.ListByUserAndBook(ctx, userID, bookID)
}
