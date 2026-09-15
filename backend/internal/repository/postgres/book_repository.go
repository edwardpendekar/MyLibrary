package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type bookRepository struct{ db *gorm.DB }

func NewBookRepository(db *gorm.DB) domain.BookRepository { return &bookRepository{db: db} }

func (r *bookRepository) Create(ctx context.Context, b *domain.Book) error {
	m := bookFromDomain(b)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	b.ID, b.CreatedAt, b.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *bookRepository) Update(ctx context.Context, b *domain.Book) error {
	m := bookFromDomain(b)
	return r.db.WithContext(ctx).Model(&bookModel{}).Where("id = ?", b.ID).Updates(map[string]interface{}{
		"slug": m.Slug, "title": m.Title, "author": m.Author, "description": m.Description,
		"language_id": m.LanguageID, "category_id": m.CategoryID, "year": m.Year, "isbn": m.ISBN,
		"cover_path": m.CoverPath, "cover_file_id": m.CoverFileID, "status": m.Status,
	}).Error
}

func (r *bookRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&bookModel{}, "id = ?", id).Error
}

func (r *bookRepository) FindByID(ctx context.Context, id int64) (*domain.Book, error) {
	var m bookModel
	err := r.db.WithContext(ctx).Preload("Language").Preload("Category").
		First(&m, "id = ? AND deleted_at IS NULL", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *bookRepository) FindBySlug(ctx context.Context, slug string) (*domain.Book, error) {
	var m bookModel
	err := r.db.WithContext(ctx).Preload("Language").Preload("Category").
		First(&m, "slug = ? AND deleted_at IS NULL", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *bookRepository) List(ctx context.Context, filter domain.BookFilter, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	limit = pagination.NormalizeLimit(limit)

	base := r.db.WithContext(ctx).Model(&bookModel{}).Preload("Language").Preload("Category").
		Where("deleted_at IS NULL")

	if filter.CategoryID != nil {
		base = base.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.LanguageID != nil {
		base = base.Where("language_id = ?", *filter.LanguageID)
	}
	if filter.Status != nil {
		base = base.Where("status = ?", *filter.Status)
	}
	if filter.Query != "" {
		base = base.Where("title ILIKE ? OR author ILIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}

	q, err := applyKeysetCursor(base, cursor, limit)
	if err != nil {
		return nil, err
	}

	var models []bookModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]domain.Book, 0, len(models))
	for _, m := range models {
		items = append(items, *m.toDomain())
	}
	hasMore, next := false, ""
	if len(items) > limit {
		items = items[:limit]
		hasMore, next = true, pagination.Encode("", items[len(items)-1].ID)
	}
	return &domain.ListResult[domain.Book]{Items: items, HasMore: hasMore, NextCursor: next}, nil
}

func (r *bookRepository) ListPopular(ctx context.Context, limit int) ([]domain.Book, error) {
	var models []bookModel
	err := r.db.WithContext(ctx).Preload("Language").Preload("Category").
		Where("deleted_at IS NULL AND status = ?", domain.BookStatusPublished).
		Order("view_count DESC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Book, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}

func (r *bookRepository) IncrementViewCount(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&bookModel{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// RecalculateCounts recomputes chapters_count/verses_count from the actual rows.
// Called once at the end of an import batch rather than via per-row triggers,
// which would otherwise slow down COPY-based bulk inserts.
func (r *bookRepository) RecalculateCounts(ctx context.Context, bookID int64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE books SET
			chapters_count = (SELECT COUNT(*) FROM chapters WHERE book_id = ?),
			verses_count   = (SELECT COUNT(*) FROM verses WHERE book_id = ?)
		WHERE id = ?`, bookID, bookID, bookID).Error
}

// FindOrCreateBySlug is safe under concurrent import jobs: it relies on the
// unique(slug) constraint via ON CONFLICT DO NOTHING rather than a racy
// check-then-insert, then re-reads the row either way.
func (r *bookRepository) FindOrCreateBySlug(ctx context.Context, b *domain.Book) (*domain.Book, bool, error) {
	m := bookFromDomain(b)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoNothing: true,
	}).Create(m)
	if result.Error != nil {
		return nil, false, result.Error
	}
	created := result.RowsAffected > 0

	existing, err := r.FindBySlug(ctx, b.Slug)
	if err != nil {
		return nil, false, err
	}
	return existing, created, nil
}
