package domain

import (
	"context"
)

// ListResult is the generic shape every cursor-paginated repository method returns.
type ListResult[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
}

type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id int64) error
	List(ctx context.Context, cursor string, limit int) (*ListResult[User], error)
	TouchLastLogin(ctx context.Context, id int64) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id int64, replacedByID *int64) error
	RevokeAllForUser(ctx context.Context, userID int64) error
}

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	ListActiveForUser(ctx context.Context, userID int64) ([]Session, error)
	Revoke(ctx context.Context, id int64) error
	Touch(ctx context.Context, id int64) error
}

type LanguageRepository interface {
	Create(ctx context.Context, l *Language) error
	Update(ctx context.Context, l *Language) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Language, error)
	FindByCode(ctx context.Context, code string) (*Language, error)
	List(ctx context.Context) ([]Language, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	List(ctx context.Context) ([]Category, error)
}

// BookFilter narrows Book listings for the Home grid and admin table.
type BookFilter struct {
	CategoryID *int64
	LanguageID *int64
	Status     *string
	Query      string // ILIKE/trigram match on title/author, used for lightweight autocomplete
}

type BookRepository interface {
	Create(ctx context.Context, b *Book) error
	Update(ctx context.Context, b *Book) error
	SoftDelete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Book, error)
	FindBySlug(ctx context.Context, slug string) (*Book, error)
	List(ctx context.Context, filter BookFilter, cursor string, limit int) (*ListResult[Book], error)
	ListPopular(ctx context.Context, limit int) ([]Book, error)
	IncrementViewCount(ctx context.Context, id int64) error
	RecalculateCounts(ctx context.Context, bookID int64) error
	FindOrCreateBySlug(ctx context.Context, b *Book) (*Book, bool, error)
}

type ChapterRepository interface {
	Create(ctx context.Context, c *Chapter) error
	FindByBookAndNumber(ctx context.Context, bookID int64, number int) (*Chapter, error)
	FindOrCreate(ctx context.Context, c *Chapter) (*Chapter, bool, error)
	ListByBook(ctx context.Context, bookID int64) ([]Chapter, error)
	FindByID(ctx context.Context, id int64) (*Chapter, error)
	RecalculateVerseCount(ctx context.Context, chapterID int64) error
}

type SectionRepository interface {
	FindOrCreate(ctx context.Context, s *Section) (*Section, bool, error)
	ListByChapter(ctx context.Context, chapterID int64) ([]Section, error)
	// UpdateEndVerse extends a section's verse range as more matching rows are
	// imported after it was first created (sections are created eagerly on their
	// first verse; later verses of the same section grow end_verse_number).
	UpdateEndVerse(ctx context.Context, sectionID int64, endVerseNumber int) error
}

type VerseRepository interface {
	FindByChapterAndNumber(ctx context.Context, chapterID int64, number int) (*Verse, error)
	ListByChapter(ctx context.Context, chapterID int64) ([]Verse, error)
	FindByID(ctx context.Context, id int64) (*Verse, error)
	// BulkUpsert inserts or updates verses in a single batch (COPY-backed for inserts).
	// Returns counts of inserted vs. updated rows for import_logs reporting.
	BulkUpsert(ctx context.Context, verses []Verse, mode string) (inserted, updated, skipped int, err error)
}

type SearchHit struct {
	BookID      int64
	BookTitle   string
	BookSlug    string
	ChapterID   *int64
	ChapterNum  *int
	VerseID     *int64
	VerseNumber *int
	Snippet     string
	Rank        float64
}

type SearchRepository interface {
	SearchVerses(ctx context.Context, query, lang string, bookID *int64, cursor string, limit int) (*ListResult[SearchHit], error)
	SearchBooks(ctx context.Context, query string, cursor string, limit int) (*ListResult[Book], error)
}

type FileRepository interface {
	Create(ctx context.Context, f *File) error
	FindByID(ctx context.Context, id int64) (*File, error)
	Delete(ctx context.Context, id int64) error
}

type PDFRepository interface {
	Upsert(ctx context.Context, p *PDF) error
	FindByBookID(ctx context.Context, bookID int64) (*PDF, error)
	// FindByBookIDs batch-loads PDFs for a page of books at once, avoiding an
	// N+1 query when rendering a book list/grid response.
	FindByBookIDs(ctx context.Context, bookIDs []int64) (map[int64]PDF, error)
}

type FavoriteRepository interface {
	Add(ctx context.Context, userID, bookID int64) error
	Remove(ctx context.Context, userID, bookID int64) error
	IsFavorite(ctx context.Context, userID, bookID int64) (bool, error)
	ListByUser(ctx context.Context, userID int64, cursor string, limit int) (*ListResult[Book], error)
}

type BookmarkRepository interface {
	Create(ctx context.Context, b *Bookmark) error
	Delete(ctx context.Context, id, userID int64) error
	ListByUserAndBook(ctx context.Context, userID, bookID int64) ([]Bookmark, error)
	UpsertLastPosition(ctx context.Context, b *Bookmark) error
	FindLastPosition(ctx context.Context, userID, bookID int64) (*Bookmark, error)
}

type NoteRepository interface {
	Create(ctx context.Context, n *Note) error
	Update(ctx context.Context, n *Note) error
	Delete(ctx context.Context, id, userID int64) error
	ListByUserAndBook(ctx context.Context, userID, bookID int64) ([]Note, error)
}

type ImportLogRepository interface {
	Create(ctx context.Context, l *ImportLog) error
	Update(ctx context.Context, l *ImportLog) error
	FindByID(ctx context.Context, id int64) (*ImportLog, error)
	List(ctx context.Context, cursor string, limit int) (*ListResult[ImportLog], error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, a *AuditLog) error
	List(ctx context.Context, cursor string, limit int) (*ListResult[AuditLog], error)
}

// DashboardStats backs the admin dashboard's summary cards.
type DashboardStats struct {
	TotalBooks       int64
	PublishedBooks   int64
	TotalChapters    int64
	TotalVerses      int64
	TotalUsers       int64
	ImportsLast30Day int64
}

type StatsRepository interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
}
