package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type importLogRepository struct{ db *gorm.DB }

func NewImportLogRepository(db *gorm.DB) domain.ImportLogRepository {
	return &importLogRepository{db: db}
}

func (r *importLogRepository) Create(ctx context.Context, l *domain.ImportLog) error {
	m := importLogFromDomain(l)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	l.ID, l.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *importLogRepository) Update(ctx context.Context, l *domain.ImportLog) error {
	m := importLogFromDomain(l)
	return r.db.WithContext(ctx).Model(&importLogModel{}).Where("id = ?", l.ID).Updates(map[string]interface{}{
		"status": m.Status, "total_rows": m.TotalRows, "processed_rows": m.ProcessedRows,
		"books_created": m.BooksCreated, "chapters_created": m.ChaptersCreated,
		"sections_created": m.SectionsCreated, "verses_inserted": m.VersesInserted,
		"verses_updated": m.VersesUpdated, "verses_skipped": m.VersesSkipped,
		"error_message": m.ErrorMessage, "started_at": m.StartedAt, "finished_at": m.FinishedAt,
	}).Error
}

func (r *importLogRepository) FindByID(ctx context.Context, id int64) (*domain.ImportLog, error) {
	var m importLogModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *importLogRepository) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.ImportLog], error) {
	limit = pagination.NormalizeLimit(limit)
	q, err := applyKeysetCursor(r.db.WithContext(ctx), cursor, limit)
	if err != nil {
		return nil, err
	}
	var models []importLogModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.ImportLog, 0, len(models))
	for _, m := range models {
		items = append(items, *m.toDomain())
	}
	hasMore, next := false, ""
	if len(items) > limit {
		items = items[:limit]
		hasMore, next = true, pagination.Encode("", items[len(items)-1].ID)
	}
	return &domain.ListResult[domain.ImportLog]{Items: items, HasMore: hasMore, NextCursor: next}, nil
}
