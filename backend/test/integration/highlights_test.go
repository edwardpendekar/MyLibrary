//go:build integration

package integration

import (
	"context"
	"testing"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/repository/postgres"
)

func TestHighlights_AddRemoveAndListByBook(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(env.DB)
	admin, err := userRepo.FindByEmail(ctx, "admin@bookreader.local")
	if err != nil || admin == nil {
		t.Fatalf("seeded admin user not found: %v", err)
	}

	bookID, chapterID := seedSearchableBook(t, ctx, env)
	verseRepo := postgres.NewVerseRepository(env.DB, env.Pool)
	verse1, err := verseRepo.FindByChapterAndNumber(ctx, chapterID, 1)
	if err != nil || verse1 == nil {
		t.Fatalf("expected seeded verse 1 to exist: %v", err)
	}
	verse2, err := verseRepo.FindByChapterAndNumber(ctx, chapterID, 2)
	if err != nil || verse2 == nil {
		t.Fatalf("expected seeded verse 2 to exist: %v", err)
	}

	highlightRepo := postgres.NewHighlightRepository(env.DB)

	if err := highlightRepo.Add(ctx, admin.ID, verse1.ID); err != nil {
		t.Fatalf("failed to add highlight: %v", err)
	}
	// Adding the same (user, verse) twice must be idempotent, not a unique
	// constraint violation — the frontend can retry a toggle without checking
	// current state first.
	if err := highlightRepo.Add(ctx, admin.ID, verse1.ID); err != nil {
		t.Fatalf("expected re-adding the same highlight to be a no-op, got: %v", err)
	}
	if err := highlightRepo.Add(ctx, admin.ID, verse2.ID); err != nil {
		t.Fatalf("failed to add second highlight: %v", err)
	}

	verseIDs, err := highlightRepo.ListVerseIDsByUserAndBook(ctx, admin.ID, bookID)
	if err != nil {
		t.Fatalf("failed to list highlights: %v", err)
	}
	if len(verseIDs) != 2 {
		t.Fatalf("expected 2 highlighted verses, got %d: %v", len(verseIDs), verseIDs)
	}

	if err := highlightRepo.Remove(ctx, admin.ID, verse1.ID); err != nil {
		t.Fatalf("failed to remove highlight: %v", err)
	}
	verseIDs, err = highlightRepo.ListVerseIDsByUserAndBook(ctx, admin.ID, bookID)
	if err != nil {
		t.Fatalf("failed to list highlights after removal: %v", err)
	}
	if len(verseIDs) != 1 || verseIDs[0] != verse2.ID {
		t.Errorf("expected only verse 2 to remain highlighted, got %v", verseIDs)
	}

	// A different, unrelated book must never show up in the list.
	otherBookID, _ := seedSearchableBookNamed(t, ctx, env, "other-book-for-highlights")
	otherVerseIDs, err := highlightRepo.ListVerseIDsByUserAndBook(ctx, admin.ID, otherBookID)
	if err != nil {
		t.Fatalf("failed to list highlights for unrelated book: %v", err)
	}
	if len(otherVerseIDs) != 0 {
		t.Errorf("expected no highlights for an unrelated book, got %v", otherVerseIDs)
	}
}

// seedSearchableBookNamed is seedSearchableBook with a caller-chosen slug, so
// a test can seed more than one book without a unique-slug collision.
func seedSearchableBookNamed(t *testing.T, ctx context.Context, env *testEnv, slug string) (bookID int64, chapterID int64) {
	t.Helper()

	bookRepo := postgres.NewBookRepository(env.DB)
	chapterRepo := postgres.NewChapterRepository(env.DB)

	book := &domain.Book{Slug: slug, Title: "Another Book", Status: domain.BookStatusPublished}
	if err := bookRepo.Create(ctx, book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}
	chapter := &domain.Chapter{BookID: book.ID, Number: 1}
	if err := chapterRepo.Create(ctx, chapter); err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}
	return book.ID, chapter.ID
}
