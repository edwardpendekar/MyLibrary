package service

import (
	"context"

	"bookreader/backend/internal/domain"
)

// AuditService is invoked by the audit middleware after every write request that
// completed successfully. Failures to write an audit row are logged, not
// propagated — an audit-log outage must never take down the actual feature.
type AuditService struct{ repo domain.AuditLogRepository }

func NewAuditService(repo domain.AuditLogRepository) *AuditService { return &AuditService{repo: repo} }

type RecordInput struct {
	UserID     *int64
	Action     string
	EntityType string
	EntityID   *string
	IPAddress  string
	UserAgent  string
	Metadata   map[string]interface{}
}

func (s *AuditService) Record(ctx context.Context, in RecordInput) error {
	return s.repo.Create(ctx, &domain.AuditLog{
		UserID: in.UserID, Action: in.Action, EntityType: in.EntityType, EntityID: in.EntityID,
		IPAddress: in.IPAddress, UserAgent: in.UserAgent, Metadata: in.Metadata,
	})
}

func (s *AuditService) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.AuditLog], error) {
	return s.repo.List(ctx, cursor, limit)
}
