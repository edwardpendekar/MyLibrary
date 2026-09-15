package postgres

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/pagination"
)

type searchRepository struct{ db *gorm.DB }

func NewSearchRepository(db *gorm.DB) domain.SearchRepository { return &searchRepository{db: db} }

type verseSearchRow struct {
	BookID      int64
	BookTitle   string
	BookSlug    string
	ChapterID   int64
	ChapterNum  int
	VerseID     int64
	VerseNumber int
	Snippet     string
	Rank        float64
}

const verseSearchSelect = `
	SELECT
		v.id AS verse_id, v.number AS verse_number, v.chapter_id, c.number AS chapter_num,
		v.book_id, b.title AS book_title, b.slug AS book_slug,
		ts_rank(v.search, q.query) AS rank,
		COALESCE(
			NULLIF(ts_headline('english', coalesce(v.text_en, ''), q.query, 'MaxWords=20,MinWords=5,HighlightAll=false'), ''),
			ts_headline('indonesian', coalesce(v.text_id, ''), q.query, 'MaxWords=20,MinWords=5,HighlightAll=false')
		) AS snippet
	FROM verses v
	JOIN chapters c ON c.id = v.chapter_id
	JOIN books b ON b.id = v.book_id AND b.deleted_at IS NULL
	CROSS JOIN (SELECT websearch_to_tsquery('english', @query) || websearch_to_tsquery('indonesian', @query) AS query) q
	WHERE v.search @@ q.query`

// SearchVerses ranks matches across both language columns using PostgreSQL FTS.
// The search column (see migration 000012) already combines an 'english'-config
// vector over text_en and an 'indonesian'-config vector over text_id, so a single
// websearch_to_tsquery per language OR'd together matches whichever language the
// verse text is actually in. Pagination is keyset on (rank, verse_id) since result
// order is by rank, not by primary key.
func (r *searchRepository) SearchVerses(ctx context.Context, query, lang string, bookID *int64, cursor string, limit int) (*domain.ListResult[domain.SearchHit], error) {
	limit = pagination.NormalizeLimit(limit)
	cur, err := pagination.Decode(cursor)
	if err != nil {
		return nil, err
	}

	sql := verseSearchSelect
	args := map[string]interface{}{"query": query, "limit": limit + 1}

	if bookID != nil {
		sql += " AND v.book_id = @bookID"
		args["bookID"] = *bookID
	}
	if cur != nil {
		cursorRank, perr := strconv.ParseFloat(cur.SortValue, 64)
		if perr != nil {
			return nil, fmt.Errorf("invalid search cursor: %w", perr)
		}
		sql += " AND (ts_rank(v.search, (SELECT websearch_to_tsquery('english', @query) || websearch_to_tsquery('indonesian', @query))), v.id) < (@cursorRank, @cursorID)"
		args["cursorRank"] = cursorRank
		args["cursorID"] = cur.ID
	}
	sql += " ORDER BY rank DESC, verse_id DESC LIMIT @limit"

	var rows []verseSearchRow
	if err := r.db.WithContext(ctx).Raw(sql, args).Scan(&rows).Error; err != nil {
		return nil, err
	}

	hits := make([]domain.SearchHit, 0, len(rows))
	for _, row := range rows {
		chID, chNum, vID, vNum := row.ChapterID, row.ChapterNum, row.VerseID, row.VerseNumber
		hits = append(hits, domain.SearchHit{
			BookID: row.BookID, BookTitle: row.BookTitle, BookSlug: row.BookSlug,
			ChapterID: &chID, ChapterNum: &chNum, VerseID: &vID, VerseNumber: &vNum,
			Snippet: row.Snippet, Rank: row.Rank,
		})
	}

	hasMore, next := false, ""
	if len(hits) > limit {
		last := rows[limit-1]
		hits = hits[:limit]
		next = pagination.Encode(strconv.FormatFloat(last.Rank, 'f', -1, 64), last.VerseID)
		hasMore = true
	}
	return &domain.ListResult[domain.SearchHit]{Items: hits, HasMore: hasMore, NextCursor: next}, nil
}

func (r *searchRepository) SearchBooks(ctx context.Context, query string, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	limit = pagination.NormalizeLimit(limit)
	cur, err := pagination.Decode(cursor)
	if err != nil {
		return nil, err
	}

	// Two-step lookup rather than "SELECT b.*, ts_rank(...) ... Scan(&[]bookSearchRow{})":
	// GORM's raw Scan does not populate an embedded struct's fields (verified: only
	// flat, non-embedded destination structs scan correctly), so step 1 fetches just
	// (id, rank) via a flat struct, and step 2 loads full Book rows by ID through the
	// normal GORM path (which also gets us Language/Category preloading for free).
	sql := `
		SELECT b.id AS id, ts_rank(b.search, websearch_to_tsquery('simple', @query)) AS rank
		FROM books b
		WHERE b.deleted_at IS NULL AND b.search @@ websearch_to_tsquery('simple', @query)`
	args := map[string]interface{}{"query": query, "limit": limit + 1}

	if cur != nil {
		cursorRank, perr := strconv.ParseFloat(cur.SortValue, 64)
		if perr != nil {
			return nil, fmt.Errorf("invalid search cursor: %w", perr)
		}
		sql += " AND (ts_rank(b.search, websearch_to_tsquery('simple', @query)), b.id) < (@cursorRank, @cursorID)"
		args["cursorRank"] = cursorRank
		args["cursorID"] = cur.ID
	}
	sql += " ORDER BY rank DESC, id DESC LIMIT @limit"

	type rankedID struct {
		ID   int64
		Rank float64
	}
	var ranked []rankedID
	if err := r.db.WithContext(ctx).Raw(sql, args).Scan(&ranked).Error; err != nil {
		return nil, err
	}

	if len(ranked) == 0 {
		return &domain.ListResult[domain.Book]{}, nil
	}

	ids := make([]int64, len(ranked))
	for i, r := range ranked {
		ids[i] = r.ID
	}
	var models []bookModel
	if err := r.db.WithContext(ctx).Preload("Language").Preload("Category").Where("id IN ?", ids).Find(&models).Error; err != nil {
		return nil, err
	}
	byID := make(map[int64]bookModel, len(models))
	for _, m := range models {
		byID[m.ID] = m
	}

	items := make([]domain.Book, 0, len(ranked))
	for _, r := range ranked {
		if m, ok := byID[r.ID]; ok {
			items = append(items, *m.toDomain())
		}
	}

	hasMore, next := false, ""
	if len(items) > limit {
		lastRank := ranked[limit-1]
		items = items[:limit]
		next = pagination.Encode(strconv.FormatFloat(lastRank.Rank, 'f', -1, 64), lastRank.ID)
		hasMore = true
	}
	return &domain.ListResult[domain.Book]{Items: items, HasMore: hasMore, NextCursor: next}, nil
}
