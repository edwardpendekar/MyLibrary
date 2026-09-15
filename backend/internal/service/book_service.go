package service

import (
	"context"

	"github.com/gosimple/slug"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
)

type BookService struct {
	books    domain.BookRepository
	chapters domain.ChapterRepository
	sections domain.SectionRepository
	verses   domain.VerseRepository
	cache    *CacheService
}

func NewBookService(
	books domain.BookRepository,
	chapters domain.ChapterRepository,
	sections domain.SectionRepository,
	verses domain.VerseRepository,
	cache *CacheService,
) *BookService {
	return &BookService{books: books, chapters: chapters, sections: sections, verses: verses, cache: cache}
}

func (s *BookService) Create(ctx context.Context, b *domain.Book, createdBy int64) error {
	if b.Slug == "" {
		b.Slug = slug.Make(b.Title)
	}
	if b.Status == "" {
		b.Status = domain.BookStatusDraft
	}
	b.CreatedBy = &createdBy
	if existing, _ := s.books.FindBySlug(ctx, b.Slug); existing != nil {
		return apperror.Conflict("a book with this slug already exists")
	}
	return s.books.Create(ctx, b)
}

func (s *BookService) Update(ctx context.Context, b *domain.Book) error {
	existing, err := s.books.FindByID(ctx, b.ID)
	if err != nil {
		return apperror.Internal("failed to load book", err)
	}
	if existing == nil {
		return apperror.NotFound("book not found")
	}
	if b.Slug == "" {
		b.Slug = existing.Slug
	}
	if err := s.books.Update(ctx, b); err != nil {
		return err
	}
	s.cache.InvalidateBook(ctx, existing.ID, existing.Slug)
	if b.Slug != existing.Slug {
		s.cache.InvalidateBook(ctx, b.ID, b.Slug)
	}
	return nil
}

func (s *BookService) Delete(ctx context.Context, id int64) error {
	existing, err := s.books.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal("failed to load book", err)
	}
	if err := s.books.SoftDelete(ctx, id); err != nil {
		return err
	}
	if existing != nil {
		s.cache.InvalidateBook(ctx, existing.ID, existing.Slug)
	}
	return nil
}

func (s *BookService) GetByID(ctx context.Context, id int64) (*domain.Book, error) {
	book, err := s.books.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("failed to load book", err)
	}
	if book == nil {
		return nil, apperror.NotFound("book not found")
	}
	return book, nil
}

// GetBySlug loads a book for the public detail page and fires an async view-count
// bump (best-effort: a lost increment under a race is an acceptable trade-off for
// not blocking the read path with a write).
func (s *BookService) GetBySlug(ctx context.Context, bookSlug string) (*domain.Book, error) {
	cacheKey := s.cache.BookDetailKey(bookSlug)
	if cached, ok := getCached[domain.Book](ctx, s.cache.c, cacheKey); ok {
		go func(id int64) {
			_ = s.books.IncrementViewCount(context.Background(), id)
		}(cached.ID)
		return cached, nil
	}

	book, err := s.books.FindBySlug(ctx, bookSlug)
	if err != nil {
		return nil, apperror.Internal("failed to load book", err)
	}
	if book == nil {
		return nil, apperror.NotFound("book not found")
	}
	setCached(ctx, s.cache.c, cacheKey, book, ttlBookDetail)

	go func(id int64) {
		_ = s.books.IncrementViewCount(context.Background(), id)
	}(book.ID)
	return book, nil
}

func (s *BookService) List(ctx context.Context, filter domain.BookFilter, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	return s.books.List(ctx, filter, cursor, limit)
}

func (s *BookService) ListPopular(ctx context.Context, limit int) ([]domain.Book, error) {
	cacheKey := s.cache.PopularBooksKey(limit)
	if cached, ok := getCached[[]domain.Book](ctx, s.cache.c, cacheKey); ok {
		return *cached, nil
	}
	books, err := s.books.ListPopular(ctx, limit)
	if err != nil {
		return nil, err
	}
	setCached(ctx, s.cache.c, cacheKey, books, ttlPopularBooks)
	return books, nil
}

func (s *BookService) ListChapters(ctx context.Context, bookID int64) ([]domain.Chapter, error) {
	return s.chapters.ListByBook(ctx, bookID)
}

// ChapterContent is the composed read-model the verse reader's center panel needs:
// the chapter, its section headings, and every verse grouped under its section.
type ChapterContent struct {
	Chapter  domain.Chapter
	Sections []domain.Section
	Verses   []domain.Verse
}

func (s *BookService) GetChapterContent(ctx context.Context, bookID int64, chapterNumber int) (*ChapterContent, error) {
	cacheKey := s.cache.ChapterContentKey(bookID, chapterNumber)
	if cached, ok := getCached[ChapterContent](ctx, s.cache.c, cacheKey); ok {
		return cached, nil
	}

	chapter, err := s.chapters.FindByBookAndNumber(ctx, bookID, chapterNumber)
	if err != nil {
		return nil, apperror.Internal("failed to load chapter", err)
	}
	if chapter == nil {
		return nil, apperror.NotFound("chapter not found")
	}

	sections, err := s.sections.ListByChapter(ctx, chapter.ID)
	if err != nil {
		return nil, apperror.Internal("failed to load sections", err)
	}
	verses, err := s.verses.ListByChapter(ctx, chapter.ID)
	if err != nil {
		return nil, apperror.Internal("failed to load verses", err)
	}

	content := &ChapterContent{Chapter: *chapter, Sections: sections, Verses: verses}
	setCached(ctx, s.cache.c, cacheKey, content, ttlChapterContent)
	return content, nil
}

func (s *BookService) GetVerse(ctx context.Context, id int64) (*domain.Verse, error) {
	verse, err := s.verses.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("failed to load verse", err)
	}
	if verse == nil {
		return nil, apperror.NotFound("verse not found")
	}
	return verse, nil
}
