package service

import (
	"context"

	"bookreader/backend/internal/domain"
)

type FavoriteService struct{ repo domain.FavoriteRepository }

func NewFavoriteService(repo domain.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

func (s *FavoriteService) Add(ctx context.Context, userID, bookID int64) error {
	return s.repo.Add(ctx, userID, bookID)
}

func (s *FavoriteService) Remove(ctx context.Context, userID, bookID int64) error {
	return s.repo.Remove(ctx, userID, bookID)
}

func (s *FavoriteService) IsFavorite(ctx context.Context, userID, bookID int64) (bool, error) {
	return s.repo.IsFavorite(ctx, userID, bookID)
}

func (s *FavoriteService) List(ctx context.Context, userID int64, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	return s.repo.ListByUser(ctx, userID, cursor, limit)
}
