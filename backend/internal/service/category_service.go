package service

import (
	"context"

	"github.com/gosimple/slug"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
)

type CategoryService struct{ repo domain.CategoryRepository }

func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) Create(ctx context.Context, c *domain.Category) error {
	if c.Slug == "" {
		c.Slug = slug.Make(c.NameEN)
	}
	if existing, _ := s.repo.FindBySlug(ctx, c.Slug); existing != nil {
		return apperror.Conflict("a category with this slug already exists")
	}
	return s.repo.Create(ctx, c)
}

func (s *CategoryService) Update(ctx context.Context, c *domain.Category) error {
	existing, err := s.repo.FindByID(ctx, c.ID)
	if err != nil {
		return apperror.Internal("failed to load category", err)
	}
	if existing == nil {
		return apperror.NotFound("category not found")
	}
	return s.repo.Update(ctx, c)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
