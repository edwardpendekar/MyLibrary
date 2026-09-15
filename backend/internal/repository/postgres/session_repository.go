package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type sessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) domain.SessionRepository { return &sessionRepository{db: db} }

func (r *sessionRepository) Create(ctx context.Context, s *domain.Session) error {
	m := &sessionModel{
		UserID: s.UserID, RefreshTokenID: s.RefreshTokenID, IPAddress: s.IPAddress,
		UserAgent: s.UserAgent, LastActiveAt: time.Now(), ExpiresAt: s.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	s.ID = m.ID
	s.CreatedAt = m.CreatedAt
	s.LastActiveAt = m.LastActiveAt
	return nil
}

func (r *sessionRepository) ListActiveForUser(ctx context.Context, userID int64) ([]domain.Session, error) {
	var models []sessionModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("last_active_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Session, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&sessionModel{}).Where("id = ?", id).Update("revoked_at", now).Error
}

func (r *sessionRepository) Touch(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&sessionModel{}).Where("id = ?", id).Update("last_active_at", time.Now()).Error
}
