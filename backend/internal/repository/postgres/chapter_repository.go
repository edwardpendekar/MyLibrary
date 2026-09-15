package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
)

type chapterRepository struct{ db *gorm.DB }

func NewChapterRepository(db *gorm.DB) domain.ChapterRepository { return &chapterRepository{db: db} }

func (r *chapterRepository) Create(ctx context.Context, c *domain.Chapter) error {
	m := chapterFromDomain(c)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	c.ID, c.CreatedAt, c.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *chapterRepository) FindByBookAndNumber(ctx context.Context, bookID int64, number int) (*domain.Chapter, error) {
	var m chapterModel
	err := r.db.WithContext(ctx).First(&m, "book_id = ? AND number = ?", bookID, number).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

// FindOrCreate relies on unique(book_id, number) via ON CONFLICT DO NOTHING,
// safe for the import pipeline processing many verses of the same chapter concurrently.
func (r *chapterRepository) FindOrCreate(ctx context.Context, c *domain.Chapter) (*domain.Chapter, bool, error) {
	m := chapterFromDomain(c)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "book_id"}, {Name: "number"}},
		DoNothing: true,
	}).Create(m)
	if result.Error != nil {
		return nil, false, result.Error
	}
	created := result.RowsAffected > 0

	existing, err := r.FindByBookAndNumber(ctx, c.BookID, c.Number)
	if err != nil {
		return nil, false, err
	}
	return existing, created, nil
}

func (r *chapterRepository) ListByBook(ctx context.Context, bookID int64) ([]domain.Chapter, error) {
	var models []chapterModel
	if err := r.db.WithContext(ctx).Where("book_id = ?", bookID).Order("number ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Chapter, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}

func (r *chapterRepository) FindByID(ctx context.Context, id int64) (*domain.Chapter, error) {
	var m chapterModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *chapterRepository) RecalculateVerseCount(ctx context.Context, chapterID int64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE chapters SET verses_count = (SELECT COUNT(*) FROM verses WHERE chapter_id = ?)
		WHERE id = ?`, chapterID, chapterID).Error
}
