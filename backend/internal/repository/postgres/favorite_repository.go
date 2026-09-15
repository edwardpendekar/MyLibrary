package postgres

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type favoriteRepository struct{ db *gorm.DB }

func NewFavoriteRepository(db *gorm.DB) domain.FavoriteRepository { return &favoriteRepository{db: db} }

func (r *favoriteRepository) Add(ctx context.Context, userID, bookID int64) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&favoriteModel{UserID: userID, BookID: bookID}).Error
}

func (r *favoriteRepository) Remove(ctx context.Context, userID, bookID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND book_id = ?", userID, bookID).
		Delete(&favoriteModel{}).Error
}

func (r *favoriteRepository) IsFavorite(ctx context.Context, userID, bookID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&favoriteModel{}).
		Where("user_id = ? AND book_id = ?", userID, bookID).Count(&count).Error
	return count > 0, err
}

func (r *favoriteRepository) ListByUser(ctx context.Context, userID int64, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	limit = pagination.NormalizeLimit(limit)
	cur, err := pagination.Decode(cursor)
	if err != nil {
		return nil, err
	}

	q := r.db.WithContext(ctx).Table("books").
		Joins("JOIN favorites ON favorites.book_id = books.id").
		Where("favorites.user_id = ? AND books.deleted_at IS NULL", userID).
		Order("books.id DESC").Limit(limit + 1)
	if cur != nil {
		q = q.Where("books.id < ?", cur.ID)
	}

	var models []bookModel
	if err := q.Select("books.*").Find(&models).Error; err != nil {
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
