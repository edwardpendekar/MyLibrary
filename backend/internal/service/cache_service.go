package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookreader/backend/pkg/cache"
)

// CacheService centralizes cache key naming and TTLs so every service caches
// consistently instead of hand-rolling Redis keys inline. All cached payloads
// are plain JSON of the same structs the DB-backed path would have returned,
// so a cache miss and a cache hit are indistinguishable to the caller.
type CacheService struct {
	c *cache.Cache
}

func NewCacheService(c *cache.Cache) *CacheService { return &CacheService{c: c} }

const (
	ttlBookDetail      = 5 * time.Minute
	ttlChapterContent  = 10 * time.Minute
	ttlPopularBooks    = 2 * time.Minute
	ttlSearchResult    = 60 * time.Second
	prefixBookDetail   = "cache:book:detail:"
	prefixChapter      = "cache:book:%d:chapter:"
	prefixPopularBooks = "cache:books:popular"
	prefixSearchVerses = "cache:search:verses:"
	prefixSearchBooks  = "cache:search:books:"
)

func getCached[T any](ctx context.Context, c *cache.Cache, key string) (*T, bool) {
	raw, ok, err := c.Get(ctx, key)
	if err != nil || !ok {
		return nil, false
	}
	var val T
	if err := json.Unmarshal([]byte(raw), &val); err != nil {
		return nil, false
	}
	return &val, true
}

func setCached(ctx context.Context, c *cache.Cache, key string, value interface{}, ttl time.Duration) {
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = c.Set(ctx, key, string(raw), ttl)
}

func (s *CacheService) BookDetailKey(slug string) string { return prefixBookDetail + slug }

func (s *CacheService) ChapterContentKey(bookID int64, chapterNumber int) string {
	return fmt.Sprintf(prefixChapter+"%d", bookID, chapterNumber)
}

func (s *CacheService) PopularBooksKey(limit int) string {
	return fmt.Sprintf("%s:%d", prefixPopularBooks, limit)
}

func (s *CacheService) SearchVersesKey(query string, bookID *int64, limit int) string {
	bookPart := "any"
	if bookID != nil {
		bookPart = fmt.Sprintf("%d", *bookID)
	}
	return fmt.Sprintf("%s%s:%s:%d", prefixSearchVerses, query, bookPart, limit)
}

func (s *CacheService) SearchBooksKey(query string, limit int) string {
	return fmt.Sprintf("%s%s:%d", prefixSearchBooks, query, limit)
}

// InvalidateBook clears everything cached for one book: its detail page, every
// cached chapter of it, and the popular-books rail (whose view counts may have
// shifted). Called after any admin write to the book or a completed import
// that touched it.
func (s *CacheService) InvalidateBook(ctx context.Context, bookID int64, slug string) {
	_ = s.c.Delete(ctx, s.BookDetailKey(slug))
	_ = s.c.DeleteByPrefix(ctx, fmt.Sprintf(prefixChapter, bookID))
	_ = s.c.DeleteByPrefix(ctx, prefixPopularBooks)
}

// InvalidateSearch drops every cached search result. Called after an import
// completes, since new/changed verses can change what any query should return.
func (s *CacheService) InvalidateSearch(ctx context.Context) {
	_ = s.c.DeleteByPrefix(ctx, prefixSearchVerses)
	_ = s.c.DeleteByPrefix(ctx, prefixSearchBooks)
}
