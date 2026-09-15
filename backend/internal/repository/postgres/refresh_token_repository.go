package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type refreshTokenRepository struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) error {
	m := &refreshTokenModel{
		UserID: t.UserID, TokenHash: t.TokenHash, CreatedByIP: t.CreatedByIP,
		UserAgent: t.UserAgent, ExpiresAt: t.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	return nil
}

func (r *refreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var m refreshTokenModel
	if err := r.db.WithContext(ctx).First(&m, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id int64, replacedByID *int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&refreshTokenModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"revoked_at":     now,
		"replaced_by_id": replacedByID,
	}).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&refreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}
