package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type categoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository { return &categoryRepository{db: db} }

func (r *categoryRepository) Create(ctx context.Context, c *domain.Category) error {
	m := categoryFromDomain(c)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	c.ID, c.CreatedAt, c.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, c *domain.Category) error {
	m := categoryFromDomain(c)
	return r.db.WithContext(ctx).Model(&categoryModel{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
		"parent_id": m.ParentID, "slug": m.Slug, "name_en": m.NameEN, "name_id": m.NameID,
		"description": m.Description,
	}).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&categoryModel{}, "id = ?", id).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	var m categoryModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	c := m.toDomain()
	return &c, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var m categoryModel
	if err := r.db.WithContext(ctx).First(&m, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	c := m.toDomain()
	return &c, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	var models []categoryModel
	if err := r.db.WithContext(ctx).Order("name_en ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Category, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}
