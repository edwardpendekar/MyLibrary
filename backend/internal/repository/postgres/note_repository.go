package postgres

import (
	"context"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type noteRepository struct{ db *gorm.DB }

func NewNoteRepository(db *gorm.DB) domain.NoteRepository { return &noteRepository{db: db} }

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	m := noteFromDomain(n)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	n.ID, n.CreatedAt, n.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *noteRepository) Update(ctx context.Context, n *domain.Note) error {
	return r.db.WithContext(ctx).Model(&noteModel{}).
		Where("id = ? AND user_id = ?", n.ID, n.UserID).
		Update("content", n.Content).Error
}

func (r *noteRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&noteModel{}).Error
}

func (r *noteRepository) ListByUserAndBook(ctx context.Context, userID, bookID int64) ([]domain.Note, error) {
	var models []noteModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND book_id = ?", userID, bookID).
		Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Note, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}
