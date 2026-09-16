//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/repository/postgres"
)

// seedSearchableBook creates a book with one chapter/section and a handful of
// verses directly through the repositories (bypassing the import pipeline,
// which is exercised separately in import_test.go) so these tests isolate the
// full-text search SQL itself: the generated tsvector columns, the GIN
// indexes, and the ts_rank-based cursor pagination in SearchRepository.
func seedSearchableBook(t *testing.T, ctx context.Context, env *testEnv) (bookID int64, chapterID int64) {
	t.Helper()

	bookRepo := postgres.NewBookRepository(env.DB)
	chapterRepo := postgres.NewChapterRepository(env.DB)
	sectionRepo := postgres.NewSectionRepository(env.DB)
	verseRepo := postgres.NewVerseRepository(env.DB, env.Pool)

	book := &domain.Book{Slug: "genesis-search-test", Title: "Genesis", Author: "Moses", Status: domain.BookStatusPublished}
	if err := bookRepo.Create(ctx, book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	chapter := &domain.Chapter{BookID: book.ID, Number: 1}
	if err := chapterRepo.Create(ctx, chapter); err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	section, _, err := sectionRepo.FindOrCreate(ctx, &domain.Section{
		BookID: book.ID, ChapterID: chapter.ID, OrderIndex: 1, StartVerseNumber: 1, EndVerseNumber: 2,
	})
	if err != nil {
		t.Fatalf("failed to create section: %v", err)
	}

	verses := []domain.Verse{
		{
			BookID: book.ID, ChapterID: chapter.ID, SectionID: &section.ID, Number: 1,
			TextEN: strPtr("In the beginning God created the heavens and the earth."),
			TextID: strPtr("Pada mulanya Allah menciptakan langit dan bumi."),
		},
		{
			BookID: book.ID, ChapterID: chapter.ID, SectionID: &section.ID, Number: 2,
			TextEN: strPtr("Now the earth was formless and empty, darkness was over the surface of the deep."),
			TextID: strPtr("Bumi belum berbentuk dan kosong, gelap gulita menutupi samudera raya."),
		},
	}
	if _, _, _, err := verseRepo.BulkUpsert(ctx, verses, domain.ImportModeInsert); err != nil {
		t.Fatalf("failed to seed verses: %v", err)
	}

	return book.ID, chapter.ID
}

func strPtr(s string) *string { return &s }

func TestSearchVerses_MatchesEnglishAndIndonesian(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()
	bookID, _ := seedSearchableBook(t, ctx, env)

	searchRepo := postgres.NewSearchRepository(env.DB)

	t.Run("english query matches text_en", func(t *testing.T) {
		result, err := searchRepo.SearchVerses(ctx, "beginning", "en", nil, "", 10)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(result.Items) != 1 {
			t.Fatalf("expected 1 hit for 'beginning', got %d", len(result.Items))
		}
		if result.Items[0].BookID != bookID {
			t.Errorf("expected hit to belong to the seeded book, got book_id=%d", result.Items[0].BookID)
		}
		if result.Items[0].VerseNumber == nil || *result.Items[0].VerseNumber != 1 {
			t.Errorf("expected the hit to be verse 1, got %+v", result.Items[0].VerseNumber)
		}
	})

	t.Run("indonesian query matches text_id", func(t *testing.T) {
		result, err := searchRepo.SearchVerses(ctx, "menciptakan", "id", nil, "", 10)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(result.Items) != 1 {
			t.Fatalf("expected 1 hit for 'menciptakan', got %d", len(result.Items))
		}
	})

	t.Run("no match returns empty result, not an error", func(t *testing.T) {
		result, err := searchRepo.SearchVerses(ctx, "xyznonexistentterm", "en", nil, "", 10)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(result.Items) != 0 {
			t.Errorf("expected no hits, got %d", len(result.Items))
		}
	})

	t.Run("bookID filter scopes results", func(t *testing.T) {
		otherBookID := bookID + 999_999 // guaranteed not to exist
		result, err := searchRepo.SearchVerses(ctx, "beginning", "en", &otherBookID, "", 10)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(result.Items) != 0 {
			t.Errorf("expected no hits when filtered to an unrelated book, got %d", len(result.Items))
		}
	})
}

func TestSearchBooks_MatchesTitleAndAuthor(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()
	seedSearchableBook(t, ctx, env)

	searchRepo := postgres.NewSearchRepository(env.DB)

	result, err := searchRepo.SearchBooks(ctx, "Genesis", "", 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 hit for title 'Genesis', got %d", len(result.Items))
	}
	if !strings.EqualFold(result.Items[0].Title, "Genesis") {
		t.Errorf("expected matched book titled Genesis, got %q", result.Items[0].Title)
	}

	byAuthor, err := searchRepo.SearchBooks(ctx, "Moses", "", 10)
	if err != nil {
		t.Fatalf("search by author failed: %v", err)
	}
	if len(byAuthor.Items) != 1 {
		t.Errorf("expected author search to also match, got %d hits", len(byAuthor.Items))
	}
}
