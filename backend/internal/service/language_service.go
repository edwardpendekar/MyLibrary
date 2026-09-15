package service

import (
	"context"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
)

type LanguageService struct{ repo domain.LanguageRepository }

func NewLanguageService(repo domain.LanguageRepository) *LanguageService {
	return &LanguageService{repo: repo}
}

func (s *LanguageService) List(ctx context.Context) ([]domain.Language, error) {
	return s.repo.List(ctx)
}

func (s *LanguageService) Create(ctx context.Context, l *domain.Language) error {
	if existing, _ := s.repo.FindByCode(ctx, l.Code); existing != nil {
		return apperror.Conflict("a language with this code already exists")
	}
	l.IsActive = true
	return s.repo.Create(ctx, l)
}

func (s *LanguageService) Update(ctx context.Context, l *domain.Language) error {
	existing, err := s.repo.FindByID(ctx, l.ID)
	if err != nil {
		return apperror.Internal("failed to load language", err)
	}
	if existing == nil {
		return apperror.NotFound("language not found")
	}
	return s.repo.Update(ctx, l)
}

func (s *LanguageService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
