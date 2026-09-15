package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
)

type pdfRepository struct{ db *gorm.DB }

func NewPDFRepository(db *gorm.DB) domain.PDFRepository { return &pdfRepository{db: db} }

func (r *pdfRepository) Upsert(ctx context.Context, p *domain.PDF) error {
	m := &pdfModel{BookID: p.BookID, FileID: p.FileID, PageCount: p.PageCount}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "book_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"file_id", "page_count", "updated_at"}),
	}).Create(m).Error
	if err != nil {
		return err
	}
	p.ID = m.ID
	return nil
}

func (r *pdfRepository) FindByBookID(ctx context.Context, bookID int64) (*domain.PDF, error) {
	var m pdfModel
	if err := r.db.WithContext(ctx).Preload("File").First(&m, "book_id = ?", bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *pdfRepository) FindByBookIDs(ctx context.Context, bookIDs []int64) (map[int64]domain.PDF, error) {
	if len(bookIDs) == 0 {
		return map[int64]domain.PDF{}, nil
	}
	var models []pdfModel
	if err := r.db.WithContext(ctx).Preload("File").Where("book_id IN ?", bookIDs).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make(map[int64]domain.PDF, len(models))
	for _, m := range models {
		out[m.BookID] = *m.toDomain()
	}
	return out, nil
}
