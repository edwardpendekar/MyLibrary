package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
)

type bookmarkRepository struct{ db *gorm.DB }

func NewBookmarkRepository(db *gorm.DB) domain.BookmarkRepository { return &bookmarkRepository{db: db} }

func (r *bookmarkRepository) Create(ctx context.Context, b *domain.Bookmark) error {
	m := bookmarkFromDomain(b)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	b.ID, b.CreatedAt, b.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *bookmarkRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&bookmarkModel{}).Error
}

func (r *bookmarkRepository) ListByUserAndBook(ctx context.Context, userID, bookID int64) ([]domain.Bookmark, error) {
	var models []bookmarkModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND book_id = ? AND is_auto = false", userID, bookID).
		Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Bookmark, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}

// UpsertLastPosition writes the single is_auto=true "remember last page" row per
// user/book, relying on the partial unique index uq_bookmarks_auto_position.
func (r *bookmarkRepository) UpsertLastPosition(ctx context.Context, b *domain.Bookmark) error {
	b.IsAuto = true
	m := bookmarkFromDomain(b)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "book_id"}},
		// TargetWhere (not Where) is what makes Postgres match this to the
		// partial unique index uq_bookmarks_auto_position — it becomes
		// `ON CONFLICT (...) WHERE is_auto = true`, the conflict arbiter's own
		// predicate, distinct from a conditional-update WHERE on DO UPDATE SET.
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "is_auto = true"}}},
		DoUpdates:   clause.AssignmentColumns([]string{"chapter_id", "verse_id", "pdf_page", "updated_at"}),
	}).Create(m).Error
}

func (r *bookmarkRepository) FindLastPosition(ctx context.Context, userID, bookID int64) (*domain.Bookmark, error) {
	var m bookmarkModel
	err := r.db.WithContext(ctx).
		First(&m, "user_id = ? AND book_id = ? AND is_auto = true", userID, bookID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}
