package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type roleRepository struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) domain.RoleRepository { return &roleRepository{db: db} }

func (r *roleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var m roleModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	role := m.toDomain()
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]domain.Role, error) {
	var models []roleModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Role, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}
