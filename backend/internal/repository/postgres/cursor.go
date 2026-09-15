package postgres

import (
	"gorm.io/gorm"

	"bookreader/backend/pkg/pagination"
)

// applyKeysetCursor implements keyset ("cursor") pagination ordered by primary
// key descending: WHERE id < :cursor.id ORDER BY id DESC LIMIT :limit+1.
// Fetching one extra row lets callers set HasMore without a second COUNT query.
func applyKeysetCursor(db *gorm.DB, cursorToken string, limit int) (*gorm.DB, error) {
	q := db.Order("id DESC").Limit(limit + 1)
	cur, err := pagination.Decode(cursorToken)
	if err != nil {
		return nil, err
	}
	if cur != nil {
		q = q.Where("id < ?", cur.ID)
	}
	return q, nil
}

// buildPage trims the "limit+1"-th row (if present) and returns the next cursor.
// ids must be the primary keys of rows, in the same order as fetched.
func buildPage(count int, limit int, lastID int64) (hasMore bool, nextCursor string) {
	if count > limit {
		return true, pagination.Encode("", lastID)
	}
	return false, ""
}
