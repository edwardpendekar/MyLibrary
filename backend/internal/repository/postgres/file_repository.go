package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type fileRepository struct{ db *gorm.DB }

func NewFileRepository(db *gorm.DB) domain.FileRepository { return &fileRepository{db: db} }

func (r *fileRepository) Create(ctx context.Context, f *domain.File) error {
	m := fileFromDomain(f)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	f.ID = m.ID
	f.CreatedAt = m.CreatedAt
	return nil
}

func (r *fileRepository) FindByID(ctx context.Context, id int64) (*domain.File, error) {
	var m fileModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *fileRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&fileModel{}, "id = ?", id).Error
}
