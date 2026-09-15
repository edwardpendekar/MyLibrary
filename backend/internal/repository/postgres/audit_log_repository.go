package postgres

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type auditLogRepository struct{ db *gorm.DB }

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository { return &auditLogRepository{db: db} }

func (r *auditLogRepository) Create(ctx context.Context, a *domain.AuditLog) error {
	metaJSON, err := json.Marshal(a.Metadata)
	if err != nil {
		return err
	}
	m := &auditLogModel{
		UserID: a.UserID, Action: a.Action, EntityType: a.EntityType, EntityID: a.EntityID,
		IPAddress: a.IPAddress, UserAgent: a.UserAgent, Metadata: metaJSON,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	a.ID, a.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.AuditLog], error) {
	limit = pagination.NormalizeLimit(limit)
	q, err := applyKeysetCursor(r.db.WithContext(ctx), cursor, limit)
	if err != nil {
		return nil, err
	}
	var models []auditLogModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.AuditLog, 0, len(models))
	for _, m := range models {
		var meta map[string]interface{}
		_ = json.Unmarshal(m.Metadata, &meta)
		items = append(items, domain.AuditLog{
			ID: m.ID, UserID: m.UserID, Action: m.Action, EntityType: m.EntityType, EntityID: m.EntityID,
			IPAddress: m.IPAddress, UserAgent: m.UserAgent, Metadata: meta, CreatedAt: m.CreatedAt,
		})
	}
	hasMore, next := false, ""
	if len(items) > limit {
		items = items[:limit]
		hasMore, next = true, pagination.Encode("", items[len(items)-1].ID)
	}
	return &domain.ListResult[domain.AuditLog]{Items: items, HasMore: hasMore, NextCursor: next}, nil
}
