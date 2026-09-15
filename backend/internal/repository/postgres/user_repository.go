package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) domain.UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	m := userFromDomain(u)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	u.ID = m.ID
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Preload("Role").First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Preload("Role").First(&m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	m := userFromDomain(u)
	return r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"role_id":           m.RoleID,
		"name":              m.Name,
		"email":             m.Email,
		"password_hash":     m.PasswordHash,
		"avatar_path":       m.AvatarPath,
		"is_active":         m.IsActive,
		"email_verified_at": m.EmailVerifiedAt,
	}).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).
		Update("deleted_at", now).Error
}

func (r *userRepository) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.User], error) {
	limit = pagination.NormalizeLimit(limit)
	q, err := applyKeysetCursor(r.db.WithContext(ctx).Preload("Role").Where("deleted_at IS NULL"), cursor, limit)
	if err != nil {
		return nil, err
	}
	var models []userModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.User, 0, len(models))
	for _, m := range models {
		items = append(items, *m.toDomain())
	}
	hasMore, next := false, ""
	if len(items) > limit {
		items = items[:limit]
		hasMore, next = true, pagination.Encode("", items[len(items)-1].ID)
	}
	return &domain.ListResult[domain.User]{Items: items, HasMore: hasMore, NextCursor: next}, nil
}

func (r *userRepository) TouchLastLogin(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).Update("last_login_at", now).Error
}
