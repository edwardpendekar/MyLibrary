package postgres

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
)

type highlightRepository struct{ db *gorm.DB }

func NewHighlightRepository(db *gorm.DB) domain.HighlightRepository {
	return &highlightRepository{db: db}
}

func (r *highlightRepository) Add(ctx context.Context, userID, verseID int64) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&highlightModel{UserID: userID, VerseID: verseID}).Error
}

func (r *highlightRepository) Remove(ctx context.Context, userID, verseID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND verse_id = ?", userID, verseID).
		Delete(&highlightModel{}).Error
}

func (r *highlightRepository) ListVerseIDsByUserAndBook(ctx context.Context, userID, bookID int64) ([]int64, error) {
	var verseIDs []int64
	err := r.db.WithContext(ctx).Table("highlights").
		Joins("JOIN verses ON verses.id = highlights.verse_id").
		Where("highlights.user_id = ? AND verses.book_id = ?", userID, bookID).
		Pluck("highlights.verse_id", &verseIDs).Error
	return verseIDs, err
}
