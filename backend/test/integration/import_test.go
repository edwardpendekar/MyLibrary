//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"bookreader/backend/internal/domain"
	"bookreader/backend/internal/repository/postgres"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/storage"
)

// importFixture wires the same repository/service graph main.go builds for
// the import pipeline, pointed at the per-test containers from setupTestEnv.
type importFixture struct {
	imports *service.ImportService
	files   *service.FileService
	verses  domain.VerseRepository
	books   domain.BookRepository
	adminID int64
}

func newImportFixture(t *testing.T, env *testEnv) *importFixture {
	t.Helper()
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(env.DB)
	admin, err := userRepo.FindByEmail(ctx, "admin@bookreader.local")
	if err != nil || admin == nil {
		t.Fatalf("seeded admin user not found: %v", err)
	}

	fileRepo := postgres.NewFileRepository(env.DB)
	pdfRepo := postgres.NewPDFRepository(env.DB)
	bookRepo := postgres.NewBookRepository(env.DB)
	chapterRepo := postgres.NewChapterRepository(env.DB)
	sectionRepo := postgres.NewSectionRepository(env.DB)
	verseRepo := postgres.NewVerseRepository(env.DB, env.Pool)
	importLogRepo := postgres.NewImportLogRepository(env.DB)

	provider := storage.NewLocalDisk(tempStorageDir(t), "http://localhost/uploads")
	cacheSvc := service.NewCacheService(env.Cache)
	fileSvc := service.NewFileService(fileRepo, pdfRepo, provider)
	importSvc := service.NewImportService(fileRepo, bookRepo, chapterRepo, sectionRepo, verseRepo, importLogRepo, provider, cacheSvc)

	return &importFixture{imports: importSvc, files: fileSvc, verses: verseRepo, books: bookRepo, adminID: admin.ID}
}

// uploadAndCommit runs the exact sequence the admin UI drives: stage the file,
// validate it (Upload), then commit it in the background and poll until the
// async run finishes — mirroring how the SSE progress endpoint would observe it.
func (f *importFixture) uploadAndCommit(t *testing.T, ctx context.Context, csvContent, mode string) *domain.ImportLog {
	t.Helper()

	uploaded, err := f.files.Upload(ctx, service.UploadInput{
		Folder: "imports", Filename: "import.csv", ContentType: "text/csv",
		Size: int64(len(csvContent)), Reader: strings.NewReader(csvContent), UploadedBy: f.adminID,
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	log, preview, err := f.imports.Upload(ctx, uploaded, f.adminID, "import.csv")
	if err != nil {
		t.Fatalf("import upload/validate failed: %v", err)
	}
	if preview == nil {
		t.Fatal("expected a non-nil preview")
	}

	if err := f.imports.Commit(log.ID, mode); err != nil {
		t.Fatalf("commit failed to start: %v", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		status, err := f.imports.GetStatus(ctx, log.ID)
		if err != nil {
			t.Fatalf("failed to poll import status: %v", err)
		}
		if status.Status == domain.ImportStatusCompleted || status.Status == domain.ImportStatusFailed {
			return status
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("import did not finish within 30s")
	return nil
}

func TestImportPipeline_InsertMode_CreatesBooksChaptersVerses(t *testing.T) {
	env := setupTestEnv(t)
	fx := newImportFixture(t, env)
	ctx := context.Background()

	csv := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1,In the beginning God created the heavens and the earth.,Pada mulanya Allah menciptakan langit dan bumi.,Creation,Penciptaan\n" +
		"Genesis,1,2,The earth was without form and void.,Bumi belum berbentuk dan kosong.,Creation,Penciptaan\n" +
		// Duplicate (chapter, verse) within the same file — insert mode must skip the repeat.
		"Genesis,1,2,DUPLICATE ROW SHOULD BE SKIPPED,DUPLIKAT,Creation,Penciptaan\n" +
		"Exodus,1,1,These are the names of the sons of Israel.,Inilah nama-nama anak Israel.,Names,Nama-nama\n"

	result := fx.uploadAndCommit(t, ctx, csv, domain.ImportModeInsert)

	if result.Status != domain.ImportStatusCompleted {
		msg := ""
		if result.ErrorMessage != nil {
			msg = *result.ErrorMessage
		}
		t.Fatalf("expected import to complete, got status=%s error=%q", result.Status, msg)
	}
	if result.BooksCreated != 2 {
		t.Errorf("expected 2 books created (Genesis, Exodus), got %d", result.BooksCreated)
	}
	if result.VersesInserted != 3 {
		t.Errorf("expected 3 verses inserted (the duplicate is skipped), got %d", result.VersesInserted)
	}
	if result.VersesSkipped != 1 {
		t.Errorf("expected 1 verse skipped (the intra-file duplicate), got %d", result.VersesSkipped)
	}

	genesis, err := fx.books.FindBySlug(ctx, "genesis")
	if err != nil || genesis == nil {
		t.Fatalf("expected Genesis to have been created: %v", err)
	}
	if genesis.VersesCount != 2 {
		t.Errorf("expected Genesis to have 2 verses after recalculation, got %d", genesis.VersesCount)
	}

	// Re-importing the exact same file in insert mode must be a full no-op:
	// every (chapter, number) already exists, so nothing new is inserted.
	second := fx.uploadAndCommit(t, ctx, csv, domain.ImportModeInsert)
	if second.VersesInserted != 0 {
		t.Errorf("expected re-import in insert mode to insert nothing, got %d inserted", second.VersesInserted)
	}
	if second.VersesSkipped != 4 {
		t.Errorf("expected re-import in insert mode to skip all 4 rows, got %d skipped", second.VersesSkipped)
	}
}

func TestImportPipeline_UpsertMode_UpdatesExistingVerses(t *testing.T) {
	env := setupTestEnv(t)
	fx := newImportFixture(t, env)
	ctx := context.Background()

	first := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1,Original text.,Teks asli.,Creation,Penciptaan\n" +
		"Genesis,1,2,Second verse.,Ayat kedua.,Creation,Penciptaan\n"
	fx.uploadAndCommit(t, ctx, first, domain.ImportModeInsert)

	// A follow-up import: verse 1 changed, verse 3 is brand new.
	second := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1,Updated text.,Teks diperbarui.,Creation,Penciptaan\n" +
		"Genesis,1,3,Third verse.,Ayat ketiga.,Creation,Penciptaan\n"
	result := fx.uploadAndCommit(t, ctx, second, domain.ImportModeUpsert)

	if result.Status != domain.ImportStatusCompleted {
		t.Fatalf("expected completion, got %s", result.Status)
	}
	if result.VersesUpdated != 1 {
		t.Errorf("expected 1 verse updated, got %d", result.VersesUpdated)
	}
	if result.VersesInserted != 1 {
		t.Errorf("expected 1 verse inserted, got %d", result.VersesInserted)
	}

	genesis, err := fx.books.FindBySlug(ctx, "genesis")
	if err != nil || genesis == nil {
		t.Fatalf("expected Genesis to exist: %v", err)
	}
	chapterRepo := postgres.NewChapterRepository(env.DB)
	chapter, err := chapterRepo.FindByBookAndNumber(ctx, genesis.ID, 1)
	if err != nil || chapter == nil {
		t.Fatalf("expected chapter 1 to exist: %v", err)
	}
	verse1, err := fx.verses.FindByChapterAndNumber(ctx, chapter.ID, 1)
	if err != nil || verse1 == nil {
		t.Fatalf("expected verse 1 to exist: %v", err)
	}
	if verse1.TextEN == nil || *verse1.TextEN != "Updated text." {
		t.Errorf("expected verse 1 text to be updated, got %v", verse1.TextEN)
	}

	if genesis.VersesCount != 3 {
		t.Errorf("expected Genesis to have 3 verses after the second import, got %d", genesis.VersesCount)
	}
}
