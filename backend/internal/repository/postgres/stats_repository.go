package postgres

import (
	"context"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type statsRepository struct{ db *gorm.DB }

func NewStatsRepository(db *gorm.DB) domain.StatsRepository { return &statsRepository{db: db} }

func (r *statsRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	var s domain.DashboardStats
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM books WHERE deleted_at IS NULL) AS total_books,
			(SELECT COUNT(*) FROM books WHERE deleted_at IS NULL AND status = 'published') AS published_books,
			(SELECT COUNT(*) FROM chapters) AS total_chapters,
			(SELECT COUNT(*) FROM verses) AS total_verses,
			(SELECT COUNT(*) FROM users WHERE deleted_at IS NULL) AS total_users,
			(SELECT COUNT(*) FROM import_logs WHERE created_at > now() - interval '30 days') AS imports_last30day
	`).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
