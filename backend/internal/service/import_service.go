package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gosimple/slug"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/storage"
)

const importBatchSize = 2000

// ImportService runs the Excel/CSV -> Book/Chapter/Section/Verse pipeline.
//
// Design trade-off, stated explicitly rather than hidden: a 1M+ row import is
// committed in many small batched transactions (one COPY + merge per
// importBatchSize rows), not one giant transaction. That is what makes 1M+ row
// throughput possible at all, but it means rollback-on-failure is a
// *compensating* action, not a database-level ROLLBACK: if a brand-new book was
// created by this run, it (and its cascaded chapters/sections/verses) is
// deleted; if the run was upserting into an already-existing book, verses
// already merged before the failure are left in place and the failure is
// reported in import_logs.error_message for manual review.
type ImportService struct {
	files      domain.FileRepository
	books      domain.BookRepository
	chapters   domain.ChapterRepository
	sections   domain.SectionRepository
	verses     domain.VerseRepository
	importLogs domain.ImportLogRepository
	storage    storage.Provider
	cache      *CacheService

	mu      sync.Mutex
	running map[int64]context.CancelFunc
}

func NewImportService(
	files domain.FileRepository,
	books domain.BookRepository,
	chapters domain.ChapterRepository,
	sections domain.SectionRepository,
	verses domain.VerseRepository,
	importLogs domain.ImportLogRepository,
	storage storage.Provider,
	cache *CacheService,
) *ImportService {
	return &ImportService{
		files: files, books: books, chapters: chapters, sections: sections, verses: verses,
		importLogs: importLogs, storage: storage, cache: cache, running: make(map[int64]context.CancelFunc),
	}
}

type PreviewResult struct {
	TotalRows      int
	DistinctBooks  int
	DistinctChapt  int
	SampleRows     []ImportRow
	ValidationErrs []RowError
}

// Upload stages the file, validates its header, and runs a full (write-free)
// scan to produce row/error counts for the admin's preview screen. It does not
// touch books/chapters/verses yet — that only happens on Commit.
func (s *ImportService) Upload(ctx context.Context, file *domain.File, uploadedBy int64, filename string) (*domain.ImportLog, *PreviewResult, error) {
	log := &domain.ImportLog{
		UploadedBy: &uploadedBy, SourceFileID: &file.ID, Filename: filename,
		Status: domain.ImportStatusValidating, Mode: domain.ImportModeInsert,
	}
	if err := s.importLogs.Create(ctx, log); err != nil {
		return nil, nil, apperror.Internal("failed to create import log", err)
	}

	localPath, cleanup, err := s.stageLocalCopy(ctx, file.StoredPath)
	if err != nil {
		s.markFailed(ctx, log, err)
		return log, nil, apperror.Validation("failed to read uploaded file", map[string]string{"_": err.Error()})
	}
	defer cleanup()

	preview, err := s.scanForPreview(localPath, filename)
	if err != nil {
		s.markFailed(ctx, log, err)
		return log, nil, apperror.Validation("file validation failed", map[string]string{"_": err.Error()})
	}

	log.Status = domain.ImportStatusReady
	log.TotalRows = preview.TotalRows
	_ = s.importLogs.Update(ctx, log)

	return log, preview, nil
}

func (s *ImportService) scanForPreview(path, filename string) (*PreviewResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader, err := NewRowReader(filepath.Ext(filename), f)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	total, _ := reader.EstimateTotal()
	result := &PreviewResult{TotalRows: total}
	books := map[string]struct{}{}
	chapters := map[string]struct{}{}
	rowCount := 0

	for {
		row, rowErr, hasMore, err := reader.Next()
		if err != nil {
			return nil, err
		}
		if !hasMore {
			break
		}
		rowCount++
		if rowErr != nil {
			if len(result.ValidationErrs) < 50 {
				result.ValidationErrs = append(result.ValidationErrs, *rowErr)
			}
			continue
		}
		books[row.Book] = struct{}{}
		chapters[fmt.Sprintf("%s|%d", row.Book, row.Chapter)] = struct{}{}
		if len(result.SampleRows) < 20 {
			result.SampleRows = append(result.SampleRows, *row)
		}
	}

	result.DistinctBooks = len(books)
	result.DistinctChapt = len(chapters)
	if total == 0 {
		result.TotalRows = rowCount
	}
	return result, nil
}

// Commit launches the actual write pipeline in the background and returns
// immediately; the admin UI polls GET /admin/import/:id (or subscribes to the
// SSE stream) for progress.
func (s *ImportService) Commit(logID int64, mode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	s.mu.Lock()
	s.running[logID] = cancel
	s.mu.Unlock()

	go func() {
		defer cancel()
		defer func() {
			s.mu.Lock()
			delete(s.running, logID)
			s.mu.Unlock()
		}()
		s.run(ctx, logID, mode)
	}()
	return nil
}

func (s *ImportService) run(ctx context.Context, logID int64, mode string) {
	log, err := s.importLogs.FindByID(ctx, logID)
	if err != nil || log == nil {
		return
	}
	if log.SourceFileID == nil {
		s.markFailed(ctx, log, fmt.Errorf("import log has no source file"))
		return
	}
	file, err := s.files.FindByID(ctx, *log.SourceFileID)
	if err != nil || file == nil {
		s.markFailed(ctx, log, fmt.Errorf("source file no longer available"))
		return
	}

	now := time.Now()
	log.Status = domain.ImportStatusImporting
	log.Mode = mode
	log.StartedAt = &now
	_ = s.importLogs.Update(ctx, log)

	localPath, cleanup, err := s.stageLocalCopy(ctx, file.StoredPath)
	if err != nil {
		s.markFailed(ctx, log, err)
		return
	}
	defer cleanup()

	newlyCreatedBookIDs, touchedBookIDs, err := s.process(ctx, localPath, file.OriginalName, log, mode)
	if err != nil {
		// Compensating rollback: only for books this run itself created.
		for _, bookID := range newlyCreatedBookIDs {
			_ = s.books.SoftDelete(ctx, bookID)
		}
		s.markFailed(ctx, log, err)
		return
	}

	// Recalculate for every book touched, not just newly created ones — an
	// import can append new chapters/verses to a book that already existed.
	for bookID := range touchedBookIDs {
		_ = s.books.RecalculateCounts(ctx, bookID)
		if book, err := s.books.FindByID(ctx, bookID); err == nil && book != nil {
			s.cache.InvalidateBook(ctx, book.ID, book.Slug)
		}
	}
	if len(touchedBookIDs) > 0 {
		s.cache.InvalidateSearch(ctx)
	}

	finished := time.Now()
	log.Status = domain.ImportStatusCompleted
	log.FinishedAt = &finished
	log.ProcessedRows = log.TotalRows
	_ = s.importLogs.Update(ctx, log)
}

type sectionState struct {
	id         int64
	titleEN    string
	titleID    string
	orderIndex int
	endVerse   int
}

// process streams the file once, materializing Books/Chapters/Sections as they
// are first encountered (cached in memory for the rest of the run, since the
// same book/chapter repeats across thousands of consecutive rows) and batching
// Verse rows into importBatchSize-sized COPY merges.
func (s *ImportService) process(ctx context.Context, path, filename string, log *domain.ImportLog, mode string) ([]int64, map[int64]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	reader, err := NewRowReader(filepath.Ext(filename), f)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	var newlyCreatedBooks []int64
	touchedBooks := map[int64]struct{}{}
	bookCache := map[string]*domain.Book{}
	chapterCache := map[string]*domain.Chapter{}

	var currentSection *sectionState
	var currentChapterID int64
	batch := make([]domain.Verse, 0, importBatchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		inserted, updated, skipped, err := s.verses.BulkUpsert(ctx, batch, mode)
		if err != nil {
			return fmt.Errorf("batch insert failed near row %d: %w", batch[len(batch)-1].Number, err)
		}
		log.VersesInserted += inserted
		log.VersesUpdated += updated
		log.VersesSkipped += skipped
		log.ProcessedRows += len(batch)
		_ = s.importLogs.Update(ctx, log)
		batch = batch[:0]
		return nil
	}

	// finalizeSection persists the accumulated end_verse_number for a section
	// we're about to stop appending to. Called once per section (not per row),
	// so it stays cheap even though verses can number in the millions.
	finalizeSection := func(sec *sectionState) {
		if sec == nil {
			return
		}
		_ = s.sections.UpdateEndVerse(ctx, sec.id, sec.endVerse)
	}

	for {
		select {
		case <-ctx.Done():
			return newlyCreatedBooks, touchedBooks, ctx.Err()
		default:
		}

		row, rowErr, hasMore, err := reader.Next()
		if err != nil {
			return newlyCreatedBooks, touchedBooks, err
		}
		if !hasMore {
			break
		}
		if rowErr != nil {
			log.VersesSkipped++
			continue
		}

		book, ok := bookCache[row.Book]
		if !ok {
			created, wasNew, err := s.findOrCreateBook(ctx, row.Book)
			if err != nil {
				return newlyCreatedBooks, touchedBooks, err
			}
			book = created
			bookCache[row.Book] = book
			if wasNew {
				newlyCreatedBooks = append(newlyCreatedBooks, book.ID)
				log.BooksCreated++
			}
		}
		touchedBooks[book.ID] = struct{}{}

		chapterKey := fmt.Sprintf("%d|%d", book.ID, row.Chapter)
		chapter, ok := chapterCache[chapterKey]
		if !ok {
			created, wasNew, err := s.chapters.FindOrCreate(ctx, &domain.Chapter{BookID: book.ID, Number: row.Chapter})
			if err != nil {
				return newlyCreatedBooks, touchedBooks, fmt.Errorf("row %d: create chapter: %w", row.RowNumber, err)
			}
			chapter = created
			chapterCache[chapterKey] = chapter
			if wasNew {
				log.ChaptersCreated++
			}
		}

		if chapter.ID != currentChapterID {
			finalizeSection(currentSection)
			currentChapterID = chapter.ID
			currentSection = nil
		}

		if currentSection == nil || currentSection.titleEN != row.TitleEN || currentSection.titleID != row.TitleID {
			finalizeSection(currentSection)
			nextOrder := 1
			if currentSection != nil {
				nextOrder = currentSection.orderIndex + 1
			}
			created, wasNew, err := s.sections.FindOrCreate(ctx, &domain.Section{
				BookID: book.ID, ChapterID: chapter.ID,
				TitleEN: strPtr(row.TitleEN), TitleID: strPtr(row.TitleID),
				OrderIndex: nextOrder, StartVerseNumber: row.Verse, EndVerseNumber: row.Verse,
			})
			if err != nil {
				return newlyCreatedBooks, touchedBooks, fmt.Errorf("row %d: create section: %w", row.RowNumber, err)
			}
			if wasNew {
				log.SectionsCreated++
			}
			currentSection = &sectionState{id: created.ID, titleEN: row.TitleEN, titleID: row.TitleID, orderIndex: nextOrder, endVerse: row.Verse}
		} else if row.Verse > currentSection.endVerse {
			currentSection.endVerse = row.Verse
		}

		sectionID := currentSection.id
		batch = append(batch, domain.Verse{
			BookID: book.ID, ChapterID: chapter.ID, SectionID: &sectionID, Number: row.Verse,
			TextEN: strPtr(row.TextEN), TextID: strPtr(row.TextID),
		})

		if len(batch) >= importBatchSize {
			if err := flush(); err != nil {
				return newlyCreatedBooks, touchedBooks, err
			}
		}
	}

	finalizeSection(currentSection)
	if err := flush(); err != nil {
		return newlyCreatedBooks, touchedBooks, err
	}
	return newlyCreatedBooks, touchedBooks, nil
}

func (s *ImportService) findOrCreateBook(ctx context.Context, title string) (*domain.Book, bool, error) {
	bookSlug := slug.Make(title)
	return s.books.FindOrCreateBySlug(ctx, &domain.Book{Slug: bookSlug, Title: title, Status: domain.BookStatusDraft})
}

func (s *ImportService) markFailed(ctx context.Context, log *domain.ImportLog, cause error) {
	msg := cause.Error()
	finished := time.Now()
	log.Status = domain.ImportStatusFailed
	log.ErrorMessage = &msg
	log.FinishedAt = &finished
	_ = s.importLogs.Update(ctx, log)
}

// stageLocalCopy copies the (possibly remote, e.g. S3) uploaded file to a local
// temp file so it can be opened twice (once to count/validate, once to stream)
// via a real io.ReadSeeker, regardless of which storage.Provider is configured.
func (s *ImportService) stageLocalCopy(ctx context.Context, storedKey string) (string, func(), error) {
	src, err := s.storage.Open(ctx, storedKey)
	if err != nil {
		return "", func() {}, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "bookreader-import-*"+filepath.Ext(storedKey))
	if err != nil {
		return "", func() {}, fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", func() {}, fmt.Errorf("stage temp file: %w", err)
	}
	tmp.Close()

	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

func (s *ImportService) GetStatus(ctx context.Context, id int64) (*domain.ImportLog, error) {
	log, err := s.importLogs.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("failed to load import log", err)
	}
	if log == nil {
		return nil, apperror.NotFound("import job not found")
	}
	return log, nil
}

func (s *ImportService) List(ctx context.Context, cursor string, limit int) (*domain.ListResult[domain.ImportLog], error) {
	return s.importLogs.List(ctx, cursor, limit)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
