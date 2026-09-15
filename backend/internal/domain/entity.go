// Package domain holds enterprise-wide entities and the repository interfaces
// that describe how they are persisted. Nothing in this package imports GORM,
// Gin, or any other framework — it is the dependency-inversion center of the
// clean architecture: repository/service/handler packages depend on domain,
// never the other way around.
package domain

import "time"

type Role struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleUser   = "user"
	RoleGuest  = "guest"
)

type User struct {
	ID              int64
	RoleID          int64
	Role            *Role
	Name            string
	Email           string
	PasswordHash    string
	AvatarPath      *string
	IsActive        bool
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type RefreshToken struct {
	ID           int64
	UserID       int64
	TokenHash    string
	ReplacedByID *int64
	CreatedByIP  string
	UserAgent    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

type Session struct {
	ID             int64
	UserID         int64
	RefreshTokenID *int64
	IPAddress      string
	UserAgent      string
	LastActiveAt   time.Time
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	CreatedAt      time.Time
}

type Language struct {
	ID         int64
	Code       string
	Name       string
	NativeName string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Category struct {
	ID          int64
	ParentID    *int64
	Slug        string
	NameEN      string
	NameID      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type File struct {
	ID             int64
	UploadedBy     *int64
	OriginalName   string
	StoredPath     string
	Provider       string
	MimeType       string
	SizeBytes      int64
	ChecksumSHA256 string
	CreatedAt      time.Time
}

const (
	BookStatusDraft     = "draft"
	BookStatusPublished = "published"
	BookStatusArchived  = "archived"
)

type Book struct {
	ID            int64
	Slug          string
	Title         string
	Author        string
	Description   string
	LanguageID    *int64
	Language      *Language
	CategoryID    *int64
	Category      *Category
	Year          *int16
	ISBN          *string
	CoverPath     *string
	CoverFileID   *int64
	Status        string
	ChaptersCount int
	VersesCount   int
	ViewCount     int64
	CreatedBy     *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

type Chapter struct {
	ID          int64
	BookID      int64
	Number      int
	TitleEN     *string
	TitleID     *string
	VersesCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Section struct {
	ID               int64
	BookID           int64
	ChapterID        int64
	TitleEN          *string
	TitleID          *string
	OrderIndex       int
	StartVerseNumber int
	EndVerseNumber   int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Verse struct {
	ID        int64
	BookID    int64
	ChapterID int64
	SectionID *int64
	Number    int
	TextEN    *string
	TextID    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PDF struct {
	ID        int64
	BookID    int64
	FileID    int64
	File      *File
	PageCount *int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Favorite struct {
	ID        int64
	UserID    int64
	BookID    int64
	CreatedAt time.Time
}

type Bookmark struct {
	ID        int64
	UserID    int64
	BookID    int64
	ChapterID *int64
	VerseID   *int64
	PDFPage   *int
	Label     *string
	IsAuto    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Note struct {
	ID        int64
	UserID    int64
	BookID    int64
	VerseID   *int64
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	ImportStatusPending    = "pending"
	ImportStatusValidating = "validating"
	ImportStatusReady      = "ready"
	ImportStatusImporting  = "importing"
	ImportStatusCompleted  = "completed"
	ImportStatusFailed     = "failed"
	ImportStatusRolledBack = "rolled_back"

	ImportModeInsert = "insert"
	ImportModeUpsert = "upsert"
)

type ImportLog struct {
	ID              int64
	UploadedBy      *int64
	SourceFileID    *int64
	Filename        string
	Status          string
	Mode            string
	TotalRows       int
	ProcessedRows   int
	BooksCreated    int
	ChaptersCreated int
	SectionsCreated int
	VersesInserted  int
	VersesUpdated   int
	VersesSkipped   int
	ErrorMessage    *string
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
}

type AuditLog struct {
	ID         int64
	UserID     *int64
	Action     string
	EntityType string
	EntityID   *string
	IPAddress  string
	UserAgent  string
	Metadata   map[string]interface{}
	CreatedAt  time.Time
}
