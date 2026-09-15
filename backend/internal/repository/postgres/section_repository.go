package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
)

type sectionRepository struct{ db *gorm.DB }

func NewSectionRepository(db *gorm.DB) domain.SectionRepository { return &sectionRepository{db: db} }

func (r *sectionRepository) FindOrCreate(ctx context.Context, s *domain.Section) (*domain.Section, bool, error) {
	m := sectionFromDomain(s)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chapter_id"}, {Name: "order_index"}},
		DoNothing: true,
	}).Create(m)
	if result.Error != nil {
		return nil, false, result.Error
	}
	created := result.RowsAffected > 0

	var found sectionModel
	err := r.db.WithContext(ctx).First(&found, "chapter_id = ? AND order_index = ?", s.ChapterID, s.OrderIndex).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return found.toDomain(), created, nil
}

func (r *sectionRepository) UpdateEndVerse(ctx context.Context, sectionID int64, endVerseNumber int) error {
	return r.db.WithContext(ctx).Model(&sectionModel{}).Where("id = ?", sectionID).
		Update("end_verse_number", endVerseNumber).Error
}

func (r *sectionRepository) ListByChapter(ctx context.Context, chapterID int64) ([]domain.Section, error) {
	var models []sectionModel
	if err := r.db.WithContext(ctx).Where("chapter_id = ?", chapterID).Order("order_index ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Section, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}
