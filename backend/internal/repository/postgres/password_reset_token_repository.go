package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type passwordResetTokenRepository struct{ db *gorm.DB }

func NewPasswordResetTokenRepository(db *gorm.DB) domain.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, t *domain.PasswordResetToken) error {
	m := &passwordResetTokenModel{UserID: t.UserID, TokenHash: t.TokenHash, ExpiresAt: t.ExpiresAt}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	return nil
}

func (r *passwordResetTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	var m passwordResetTokenModel
	if err := r.db.WithContext(ctx).First(&m, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *passwordResetTokenRepository) MarkUsed(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&passwordResetTokenModel{}).Where("id = ?", id).Update("used_at", now).Error
}

func (r *passwordResetTokenRepository) InvalidateAllForUser(ctx context.Context, userID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&passwordResetTokenModel{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}
