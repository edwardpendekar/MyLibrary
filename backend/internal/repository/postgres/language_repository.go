package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type languageRepository struct{ db *gorm.DB }

func NewLanguageRepository(db *gorm.DB) domain.LanguageRepository { return &languageRepository{db: db} }

func (r *languageRepository) Create(ctx context.Context, l *domain.Language) error {
	m := languageFromDomain(l)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	l.ID, l.CreatedAt, l.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *languageRepository) Update(ctx context.Context, l *domain.Language) error {
	m := languageFromDomain(l)
	return r.db.WithContext(ctx).Model(&languageModel{}).Where("id = ?", l.ID).Updates(map[string]interface{}{
		"code": m.Code, "name": m.Name, "native_name": m.NativeName, "is_active": m.IsActive,
	}).Error
}

func (r *languageRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&languageModel{}, "id = ?", id).Error
}

func (r *languageRepository) FindByID(ctx context.Context, id int64) (*domain.Language, error) {
	var m languageModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	l := m.toDomain()
	return &l, nil
}

func (r *languageRepository) FindByCode(ctx context.Context, code string) (*domain.Language, error) {
	var m languageModel
	if err := r.db.WithContext(ctx).First(&m, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	l := m.toDomain()
	return &l, nil
}

func (r *languageRepository) List(ctx context.Context) ([]domain.Language, error) {
	var models []languageModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Language, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}
