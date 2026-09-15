package service

import (
	"context"

	"bookreader/backend/internal/domain"
)

type StatsService struct{ repo domain.StatsRepository }

func NewStatsService(repo domain.StatsRepository) *StatsService { return &StatsService{repo: repo} }

func (s *StatsService) Dashboard(ctx context.Context) (*domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}
