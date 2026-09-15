package service

import (
	"context"

	"bookreader/backend/internal/domain"
	"bookreader/backend/pkg/apperror"
)

// SearchService caches only the first page (no cursor) of each distinct query —
// that is the overwhelming majority of real search traffic ("popular queries"
// per the spec) and keeps the cache from growing unbounded with every possible
// pagination cursor.
type SearchService struct {
	repo  domain.SearchRepository
	cache *CacheService
}

func NewSearchService(repo domain.SearchRepository, cache *CacheService) *SearchService {
	return &SearchService{repo: repo, cache: cache}
}

func (s *SearchService) SearchVerses(ctx context.Context, query, lang string, bookID *int64, cursor string, limit int) (*domain.ListResult[domain.SearchHit], error) {
	if query == "" {
		return nil, apperror.Validation("q is required", map[string]string{"q": "is required"})
	}

	cacheable := cursor == ""
	var cacheKey string
	if cacheable {
		cacheKey = s.cache.SearchVersesKey(query, bookID, limit)
		if cached, ok := getCached[domain.ListResult[domain.SearchHit]](ctx, s.cache.c, cacheKey); ok {
			return cached, nil
		}
	}

	result, err := s.repo.SearchVerses(ctx, query, lang, bookID, cursor, limit)
	if err != nil {
		return nil, apperror.Internal("search failed", err)
	}
	if cacheable {
		setCached(ctx, s.cache.c, cacheKey, result, ttlSearchResult)
	}
	return result, nil
}

func (s *SearchService) SearchBooks(ctx context.Context, query, cursor string, limit int) (*domain.ListResult[domain.Book], error) {
	if query == "" {
		return nil, apperror.Validation("q is required", map[string]string{"q": "is required"})
	}

	cacheable := cursor == ""
	var cacheKey string
	if cacheable {
		cacheKey = s.cache.SearchBooksKey(query, limit)
		if cached, ok := getCached[domain.ListResult[domain.Book]](ctx, s.cache.c, cacheKey); ok {
			return cached, nil
		}
	}

	result, err := s.repo.SearchBooks(ctx, query, cursor, limit)
	if err != nil {
		return nil, apperror.Internal("search failed", err)
	}
	if cacheable {
		setCached(ctx, s.cache.c, cacheKey, result, ttlSearchResult)
	}
	return result, nil
}
