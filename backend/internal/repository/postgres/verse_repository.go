package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
)

type verseRepository struct {
	db   *gorm.DB
	pool *pgxpool.Pool // raw pgx pool, needed for CopyFrom (not exposed via database/sql)
}

func NewVerseRepository(db *gorm.DB, pool *pgxpool.Pool) domain.VerseRepository {
	return &verseRepository{db: db, pool: pool}
}

func (r *verseRepository) FindByChapterAndNumber(ctx context.Context, chapterID int64, number int) (*domain.Verse, error) {
	var m verseModel
	err := r.db.WithContext(ctx).First(&m, "chapter_id = ? AND number = ?", chapterID, number).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *verseRepository) ListByChapter(ctx context.Context, chapterID int64) ([]domain.Verse, error) {
	var models []verseModel
	if err := r.db.WithContext(ctx).Where("chapter_id = ?", chapterID).Order("number ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Verse, len(models))
	for i, m := range models {
		out[i] = *m.toDomain()
	}
	return out, nil
}

func (r *verseRepository) FindByID(ctx context.Context, id int64) (*domain.Verse, error) {
	var m verseModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

// BulkUpsert is the throughput-critical path for Excel imports of 1M+ verses.
// Strategy: COPY the batch into a per-transaction TEMP TABLE (the pgx wire-protocol
// COPY, orders of magnitude faster than row-by-row INSERTs), then merge it into
// `verses` with a single set-based INSERT ... SELECT ... ON CONFLICT. This keeps
// per-batch cost to two round trips regardless of batch size.
func (r *verseRepository) BulkUpsert(ctx context.Context, verses []domain.Verse, mode string) (inserted, updated, skipped int, err error) {
	if len(verses) == 0 {
		return 0, 0, 0, nil
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after a successful Commit

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_verse_import (
			book_id    BIGINT,
			chapter_id BIGINT,
			section_id BIGINT,
			number     INTEGER,
			text_en    TEXT,
			text_id    TEXT
		) ON COMMIT DROP`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("create staging table: %w", err)
	}

	rows := make([][]interface{}, len(verses))
	for i, v := range verses {
		rows[i] = []interface{}{v.BookID, v.ChapterID, v.SectionID, v.Number, v.TextEN, v.TextID}
	}
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"tmp_verse_import"},
		[]string{"book_id", "chapter_id", "section_id", "number", "text_en", "text_id"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("copy staging rows: %w", err)
	}

	var mergeSQL string
	if mode == domain.ImportModeUpsert {
		mergeSQL = `
			INSERT INTO verses (book_id, chapter_id, section_id, number, text_en, text_id)
			SELECT book_id, chapter_id, section_id, number, text_en, text_id FROM tmp_verse_import
			ON CONFLICT (chapter_id, number) DO UPDATE SET
				text_en = EXCLUDED.text_en,
				text_id = EXCLUDED.text_id,
				section_id = EXCLUDED.section_id,
				updated_at = now()
			RETURNING (xmax = 0) AS inserted`
	} else {
		mergeSQL = `
			INSERT INTO verses (book_id, chapter_id, section_id, number, text_en, text_id)
			SELECT book_id, chapter_id, section_id, number, text_en, text_id FROM tmp_verse_import
			ON CONFLICT (chapter_id, number) DO NOTHING
			RETURNING true AS inserted`
	}

	result, err := tx.Query(ctx, mergeSQL)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("merge staging rows: %w", err)
	}
	for result.Next() {
		var wasInserted bool
		if scanErr := result.Scan(&wasInserted); scanErr != nil {
			result.Close()
			return 0, 0, 0, fmt.Errorf("scan merge result: %w", scanErr)
		}
		if wasInserted {
			inserted++
		} else {
			updated++
		}
	}
	result.Close()
	if err := result.Err(); err != nil {
		return 0, 0, 0, fmt.Errorf("iterate merge result: %w", err)
	}
	skipped = len(verses) - inserted - updated

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, 0, fmt.Errorf("commit tx: %w", err)
	}
	return inserted, updated, skipped, nil
}
