package postgres

import (
	"time"

	"gorm.io/datatypes"

	"bookreader/backend/internal/domain"
)

// GORM persistence models. Kept separate from domain entities on purpose: domain
// stays free of ORM tags, and a schema/column rename here never leaks into
// business logic — only the toDomain/fromXxx mapping functions change.

type roleModel struct {
	ID          int64     `gorm:"primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (roleModel) TableName() string { return "roles" }

func (m roleModel) toDomain() domain.Role {
	return domain.Role{ID: m.ID, Name: m.Name, Description: m.Description, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

type userModel struct {
	ID              int64      `gorm:"primaryKey"`
	RoleID          int64      `gorm:"column:role_id"`
	Role            *roleModel `gorm:"foreignKey:RoleID"`
	Name            string     `gorm:"column:name"`
	Email           string     `gorm:"column:email"`
	PasswordHash    string     `gorm:"column:password_hash"`
	AvatarPath      *string    `gorm:"column:avatar_path"`
	IsActive        bool       `gorm:"column:is_active"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (userModel) TableName() string { return "users" }

func (m userModel) toDomain() *domain.User {
	u := &domain.User{
		ID: m.ID, RoleID: m.RoleID, Name: m.Name, Email: m.Email, PasswordHash: m.PasswordHash,
		AvatarPath: m.AvatarPath, IsActive: m.IsActive, EmailVerifiedAt: m.EmailVerifiedAt,
		LastLoginAt: m.LastLoginAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: m.DeletedAt,
	}
	if m.Role != nil {
		role := m.Role.toDomain()
		u.Role = &role
	}
	return u
}

func userFromDomain(u *domain.User) *userModel {
	return &userModel{
		ID: u.ID, RoleID: u.RoleID, Name: u.Name, Email: u.Email, PasswordHash: u.PasswordHash,
		AvatarPath: u.AvatarPath, IsActive: u.IsActive, EmailVerifiedAt: u.EmailVerifiedAt,
		LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, DeletedAt: u.DeletedAt,
	}
}

type refreshTokenModel struct {
	ID           int64      `gorm:"primaryKey"`
	UserID       int64      `gorm:"column:user_id"`
	TokenHash    string     `gorm:"column:token_hash"`
	ReplacedByID *int64     `gorm:"column:replaced_by_id"`
	CreatedByIP  string     `gorm:"column:created_by_ip"`
	UserAgent    string     `gorm:"column:user_agent"`
	ExpiresAt    time.Time  `gorm:"column:expires_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }

func (m refreshTokenModel) toDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID: m.ID, UserID: m.UserID, TokenHash: m.TokenHash, ReplacedByID: m.ReplacedByID,
		CreatedByIP: m.CreatedByIP, UserAgent: m.UserAgent, ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt,
	}
}

type passwordResetTokenModel struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (passwordResetTokenModel) TableName() string { return "password_reset_tokens" }

func (m passwordResetTokenModel) toDomain() *domain.PasswordResetToken {
	return &domain.PasswordResetToken{
		ID: m.ID, UserID: m.UserID, TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt, UsedAt: m.UsedAt, CreatedAt: m.CreatedAt,
	}
}

type sessionModel struct {
	ID             int64      `gorm:"primaryKey"`
	UserID         int64      `gorm:"column:user_id"`
	RefreshTokenID *int64     `gorm:"column:refresh_token_id"`
	IPAddress      string     `gorm:"column:ip_address"`
	UserAgent      string     `gorm:"column:user_agent"`
	LastActiveAt   time.Time  `gorm:"column:last_active_at"`
	ExpiresAt      time.Time  `gorm:"column:expires_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
}

func (sessionModel) TableName() string { return "sessions" }

func (m sessionModel) toDomain() domain.Session {
	return domain.Session{
		ID: m.ID, UserID: m.UserID, RefreshTokenID: m.RefreshTokenID, IPAddress: m.IPAddress,
		UserAgent: m.UserAgent, LastActiveAt: m.LastActiveAt, ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt,
	}
}

type languageModel struct {
	ID         int64     `gorm:"primaryKey"`
	Code       string    `gorm:"column:code"`
	Name       string    `gorm:"column:name"`
	NativeName string    `gorm:"column:native_name"`
	IsActive   bool      `gorm:"column:is_active"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (languageModel) TableName() string { return "languages" }

func (m languageModel) toDomain() domain.Language {
	return domain.Language{
		ID: m.ID, Code: m.Code, Name: m.Name, NativeName: m.NativeName,
		IsActive: m.IsActive, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func languageFromDomain(l *domain.Language) *languageModel {
	return &languageModel{
		ID: l.ID, Code: l.Code, Name: l.Name, NativeName: l.NativeName,
		IsActive: l.IsActive, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt,
	}
}

type categoryModel struct {
	ID          int64     `gorm:"primaryKey"`
	ParentID    *int64    `gorm:"column:parent_id"`
	Slug        string    `gorm:"column:slug"`
	NameEN      string    `gorm:"column:name_en"`
	NameID      string    `gorm:"column:name_id"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (categoryModel) TableName() string { return "categories" }

func (m categoryModel) toDomain() domain.Category {
	return domain.Category{
		ID: m.ID, ParentID: m.ParentID, Slug: m.Slug, NameEN: m.NameEN, NameID: m.NameID,
		Description: m.Description, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func categoryFromDomain(c *domain.Category) *categoryModel {
	return &categoryModel{
		ID: c.ID, ParentID: c.ParentID, Slug: c.Slug, NameEN: c.NameEN, NameID: c.NameID,
		Description: c.Description, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

type fileModel struct {
	ID             int64     `gorm:"primaryKey"`
	UploadedBy     *int64    `gorm:"column:uploaded_by"`
	OriginalName   string    `gorm:"column:original_name"`
	StoredPath     string    `gorm:"column:stored_path"`
	Provider       string    `gorm:"column:provider"`
	MimeType       string    `gorm:"column:mime_type"`
	SizeBytes      int64     `gorm:"column:size_bytes"`
	ChecksumSHA256 string    `gorm:"column:checksum_sha256"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (fileModel) TableName() string { return "files" }

func (m fileModel) toDomain() *domain.File {
	return &domain.File{
		ID: m.ID, UploadedBy: m.UploadedBy, OriginalName: m.OriginalName, StoredPath: m.StoredPath,
		Provider: m.Provider, MimeType: m.MimeType, SizeBytes: m.SizeBytes,
		ChecksumSHA256: m.ChecksumSHA256, CreatedAt: m.CreatedAt,
	}
}

func fileFromDomain(f *domain.File) *fileModel {
	return &fileModel{
		ID: f.ID, UploadedBy: f.UploadedBy, OriginalName: f.OriginalName, StoredPath: f.StoredPath,
		Provider: f.Provider, MimeType: f.MimeType, SizeBytes: f.SizeBytes,
		ChecksumSHA256: f.ChecksumSHA256, CreatedAt: f.CreatedAt,
	}
}

type bookModel struct {
	ID            int64          `gorm:"primaryKey"`
	Slug          string         `gorm:"column:slug"`
	Title         string         `gorm:"column:title"`
	Author        string         `gorm:"column:author"`
	Description   string         `gorm:"column:description"`
	LanguageID    *int64         `gorm:"column:language_id"`
	Language      *languageModel `gorm:"foreignKey:LanguageID"`
	CategoryID    *int64         `gorm:"column:category_id"`
	Category      *categoryModel `gorm:"foreignKey:CategoryID"`
	Year          *int16         `gorm:"column:year"`
	ISBN          *string        `gorm:"column:isbn"`
	CoverPath     *string        `gorm:"column:cover_path"`
	CoverFileID   *int64         `gorm:"column:cover_file_id"`
	Status        string         `gorm:"column:status"`
	ChaptersCount int            `gorm:"column:chapters_count"`
	VersesCount   int            `gorm:"column:verses_count"`
	ViewCount     int64          `gorm:"column:view_count"`
	CreatedBy     *int64         `gorm:"column:created_by"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     *time.Time     `gorm:"column:deleted_at"`
}

func (bookModel) TableName() string { return "books" }

func (m bookModel) toDomain() *domain.Book {
	b := &domain.Book{
		ID: m.ID, Slug: m.Slug, Title: m.Title, Author: m.Author, Description: m.Description,
		LanguageID: m.LanguageID, CategoryID: m.CategoryID, Year: m.Year, ISBN: m.ISBN,
		CoverPath: m.CoverPath, CoverFileID: m.CoverFileID, Status: m.Status,
		ChaptersCount: m.ChaptersCount, VersesCount: m.VersesCount, ViewCount: m.ViewCount,
		CreatedBy: m.CreatedBy, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: m.DeletedAt,
	}
	if m.Language != nil {
		lang := m.Language.toDomain()
		b.Language = &lang
	}
	if m.Category != nil {
		cat := m.Category.toDomain()
		b.Category = &cat
	}
	return b
}

func bookFromDomain(b *domain.Book) *bookModel {
	return &bookModel{
		ID: b.ID, Slug: b.Slug, Title: b.Title, Author: b.Author, Description: b.Description,
		LanguageID: b.LanguageID, CategoryID: b.CategoryID, Year: b.Year, ISBN: b.ISBN,
		CoverPath: b.CoverPath, CoverFileID: b.CoverFileID, Status: b.Status,
		ChaptersCount: b.ChaptersCount, VersesCount: b.VersesCount, ViewCount: b.ViewCount,
		CreatedBy: b.CreatedBy, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt, DeletedAt: b.DeletedAt,
	}
}

type chapterModel struct {
	ID          int64     `gorm:"primaryKey"`
	BookID      int64     `gorm:"column:book_id"`
	Number      int       `gorm:"column:number"`
	TitleEN     *string   `gorm:"column:title_en"`
	TitleID     *string   `gorm:"column:title_id"`
	VersesCount int       `gorm:"column:verses_count"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (chapterModel) TableName() string { return "chapters" }

func (m chapterModel) toDomain() *domain.Chapter {
	return &domain.Chapter{
		ID: m.ID, BookID: m.BookID, Number: m.Number, TitleEN: m.TitleEN, TitleID: m.TitleID,
		VersesCount: m.VersesCount, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func chapterFromDomain(c *domain.Chapter) *chapterModel {
	return &chapterModel{
		ID: c.ID, BookID: c.BookID, Number: c.Number, TitleEN: c.TitleEN, TitleID: c.TitleID,
		VersesCount: c.VersesCount, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

type sectionModel struct {
	ID               int64     `gorm:"primaryKey"`
	BookID           int64     `gorm:"column:book_id"`
	ChapterID        int64     `gorm:"column:chapter_id"`
	TitleEN          *string   `gorm:"column:title_en"`
	TitleID          *string   `gorm:"column:title_id"`
	OrderIndex       int       `gorm:"column:order_index"`
	StartVerseNumber int       `gorm:"column:start_verse_number"`
	EndVerseNumber   int       `gorm:"column:end_verse_number"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (sectionModel) TableName() string { return "sections" }

func (m sectionModel) toDomain() *domain.Section {
	return &domain.Section{
		ID: m.ID, BookID: m.BookID, ChapterID: m.ChapterID, TitleEN: m.TitleEN, TitleID: m.TitleID,
		OrderIndex: m.OrderIndex, StartVerseNumber: m.StartVerseNumber, EndVerseNumber: m.EndVerseNumber,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func sectionFromDomain(s *domain.Section) *sectionModel {
	return &sectionModel{
		ID: s.ID, BookID: s.BookID, ChapterID: s.ChapterID, TitleEN: s.TitleEN, TitleID: s.TitleID,
		OrderIndex: s.OrderIndex, StartVerseNumber: s.StartVerseNumber, EndVerseNumber: s.EndVerseNumber,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

type verseModel struct {
	ID        int64     `gorm:"primaryKey"`
	BookID    int64     `gorm:"column:book_id"`
	ChapterID int64     `gorm:"column:chapter_id"`
	SectionID *int64    `gorm:"column:section_id"`
	Number    int       `gorm:"column:number"`
	TextEN    *string   `gorm:"column:text_en"`
	TextID    *string   `gorm:"column:text_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (verseModel) TableName() string { return "verses" }

func (m verseModel) toDomain() *domain.Verse {
	return &domain.Verse{
		ID: m.ID, BookID: m.BookID, ChapterID: m.ChapterID, SectionID: m.SectionID, Number: m.Number,
		TextEN: m.TextEN, TextID: m.TextID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

type pdfModel struct {
	ID        int64      `gorm:"primaryKey"`
	BookID    int64      `gorm:"column:book_id"`
	FileID    int64      `gorm:"column:file_id"`
	File      *fileModel `gorm:"foreignKey:FileID"`
	PageCount *int       `gorm:"column:page_count"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (pdfModel) TableName() string { return "pdfs" }

func (m pdfModel) toDomain() *domain.PDF {
	p := &domain.PDF{ID: m.ID, BookID: m.BookID, FileID: m.FileID, PageCount: m.PageCount, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
	if m.File != nil {
		p.File = m.File.toDomain()
	}
	return p
}

type favoriteModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	BookID    int64     `gorm:"column:book_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (favoriteModel) TableName() string { return "favorites" }

type highlightModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	VerseID   int64     `gorm:"column:verse_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (highlightModel) TableName() string { return "highlights" }

type bookmarkModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	BookID    int64     `gorm:"column:book_id"`
	ChapterID *int64    `gorm:"column:chapter_id"`
	VerseID   *int64    `gorm:"column:verse_id"`
	PDFPage   *int      `gorm:"column:pdf_page"`
	Label     *string   `gorm:"column:label"`
	IsAuto    bool      `gorm:"column:is_auto"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (bookmarkModel) TableName() string { return "bookmarks" }

func (m bookmarkModel) toDomain() *domain.Bookmark {
	return &domain.Bookmark{
		ID: m.ID, UserID: m.UserID, BookID: m.BookID, ChapterID: m.ChapterID, VerseID: m.VerseID,
		PDFPage: m.PDFPage, Label: m.Label, IsAuto: m.IsAuto, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func bookmarkFromDomain(b *domain.Bookmark) *bookmarkModel {
	return &bookmarkModel{
		ID: b.ID, UserID: b.UserID, BookID: b.BookID, ChapterID: b.ChapterID, VerseID: b.VerseID,
		PDFPage: b.PDFPage, Label: b.Label, IsAuto: b.IsAuto, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

type noteModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	BookID    int64     `gorm:"column:book_id"`
	VerseID   *int64    `gorm:"column:verse_id"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (noteModel) TableName() string { return "notes" }

func (m noteModel) toDomain() *domain.Note {
	return &domain.Note{
		ID: m.ID, UserID: m.UserID, BookID: m.BookID, VerseID: m.VerseID,
		Content: m.Content, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func noteFromDomain(n *domain.Note) *noteModel {
	return &noteModel{
		ID: n.ID, UserID: n.UserID, BookID: n.BookID, VerseID: n.VerseID,
		Content: n.Content, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}
}

type importLogModel struct {
	ID              int64      `gorm:"primaryKey"`
	UploadedBy      *int64     `gorm:"column:uploaded_by"`
	SourceFileID    *int64     `gorm:"column:source_file_id"`
	Filename        string     `gorm:"column:filename"`
	Status          string     `gorm:"column:status"`
	Mode            string     `gorm:"column:mode"`
	TotalRows       int        `gorm:"column:total_rows"`
	ProcessedRows   int        `gorm:"column:processed_rows"`
	BooksCreated    int        `gorm:"column:books_created"`
	ChaptersCreated int        `gorm:"column:chapters_created"`
	SectionsCreated int        `gorm:"column:sections_created"`
	VersesInserted  int        `gorm:"column:verses_inserted"`
	VersesUpdated   int        `gorm:"column:verses_updated"`
	VersesSkipped   int        `gorm:"column:verses_skipped"`
	ErrorMessage    *string    `gorm:"column:error_message"`
	StartedAt       *time.Time `gorm:"column:started_at"`
	FinishedAt      *time.Time `gorm:"column:finished_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
}

func (importLogModel) TableName() string { return "import_logs" }

func (m importLogModel) toDomain() *domain.ImportLog {
	return &domain.ImportLog{
		ID: m.ID, UploadedBy: m.UploadedBy, SourceFileID: m.SourceFileID, Filename: m.Filename,
		Status: m.Status, Mode: m.Mode, TotalRows: m.TotalRows, ProcessedRows: m.ProcessedRows,
		BooksCreated: m.BooksCreated, ChaptersCreated: m.ChaptersCreated, SectionsCreated: m.SectionsCreated,
		VersesInserted: m.VersesInserted, VersesUpdated: m.VersesUpdated, VersesSkipped: m.VersesSkipped,
		ErrorMessage: m.ErrorMessage, StartedAt: m.StartedAt, FinishedAt: m.FinishedAt, CreatedAt: m.CreatedAt,
	}
}

func importLogFromDomain(l *domain.ImportLog) *importLogModel {
	return &importLogModel{
		ID: l.ID, UploadedBy: l.UploadedBy, SourceFileID: l.SourceFileID, Filename: l.Filename,
		Status: l.Status, Mode: l.Mode, TotalRows: l.TotalRows, ProcessedRows: l.ProcessedRows,
		BooksCreated: l.BooksCreated, ChaptersCreated: l.ChaptersCreated, SectionsCreated: l.SectionsCreated,
		VersesInserted: l.VersesInserted, VersesUpdated: l.VersesUpdated, VersesSkipped: l.VersesSkipped,
		ErrorMessage: l.ErrorMessage, StartedAt: l.StartedAt, FinishedAt: l.FinishedAt, CreatedAt: l.CreatedAt,
	}
}

type auditLogModel struct {
	ID         int64          `gorm:"primaryKey"`
	UserID     *int64         `gorm:"column:user_id"`
	Action     string         `gorm:"column:action"`
	EntityType string         `gorm:"column:entity_type"`
	EntityID   *string        `gorm:"column:entity_id"`
	IPAddress  string         `gorm:"column:ip_address"`
	UserAgent  string         `gorm:"column:user_agent"`
	Metadata   datatypes.JSON `gorm:"column:metadata"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
}

func (auditLogModel) TableName() string { return "audit_logs" }
