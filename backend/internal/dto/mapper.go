package dto

import "bookreader/backend/internal/domain"

func ToUserResponse(u *domain.User) UserResponse {
	role := domain.RoleUser
	if u.Role != nil {
		role = u.Role.Name
	}
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: role, IsActive: u.IsActive, CreatedAt: u.CreatedAt}
}

func ToLanguageResponse(l *domain.Language) *LanguageResponse {
	if l == nil {
		return nil
	}
	return &LanguageResponse{ID: l.ID, Code: l.Code, Name: l.Name, NativeName: l.NativeName}
}

func ToCategoryResponse(c *domain.Category) *CategoryResponse {
	if c == nil {
		return nil
	}
	return &CategoryResponse{ID: c.ID, Slug: c.Slug, NameEN: c.NameEN, NameID: c.NameID}
}

func ToChapterResponse(c *domain.Chapter) ChapterResponse {
	return ChapterResponse{ID: c.ID, Number: c.Number, TitleEN: c.TitleEN, TitleID: c.TitleID, VersesCount: c.VersesCount}
}

func ToSectionResponse(s *domain.Section) SectionResponse {
	return SectionResponse{
		ID: s.ID, TitleEN: s.TitleEN, TitleID: s.TitleID,
		StartVerseNumber: s.StartVerseNumber, EndVerseNumber: s.EndVerseNumber,
	}
}

func ToVerseResponse(v *domain.Verse) VerseResponse {
	return VerseResponse{ID: v.ID, Number: v.Number, TextEN: v.TextEN, TextID: v.TextID, SectionID: v.SectionID}
}

func ToBookmarkResponse(b *domain.Bookmark) BookmarkResponse {
	return BookmarkResponse{
		ID: b.ID, BookID: b.BookID, ChapterID: b.ChapterID, VerseID: b.VerseID,
		PDFPage: b.PDFPage, Label: b.Label, CreatedAt: b.CreatedAt,
	}
}

func ToDashboardStatsResponse(s *domain.DashboardStats) DashboardStatsResponse {
	return DashboardStatsResponse{
		TotalBooks: s.TotalBooks, PublishedBooks: s.PublishedBooks, TotalChapters: s.TotalChapters,
		TotalVerses: s.TotalVerses, TotalUsers: s.TotalUsers, ImportsLast30Day: s.ImportsLast30Day,
	}
}

func ToAuditLogResponse(a *domain.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID: a.ID, UserID: a.UserID, Action: a.Action, EntityType: a.EntityType, EntityID: a.EntityID,
		IPAddress: a.IPAddress, UserAgent: a.UserAgent, Metadata: a.Metadata, CreatedAt: a.CreatedAt,
	}
}

func ToImportLogResponse(l *domain.ImportLog) ImportLogResponse {
	return ImportLogResponse{
		ID: l.ID, Filename: l.Filename, Status: l.Status, Mode: l.Mode,
		TotalRows: l.TotalRows, ProcessedRows: l.ProcessedRows,
		BooksCreated: l.BooksCreated, ChaptersCreated: l.ChaptersCreated, SectionsCreated: l.SectionsCreated,
		VersesInserted: l.VersesInserted, VersesUpdated: l.VersesUpdated, VersesSkipped: l.VersesSkipped,
		ErrorMessage: l.ErrorMessage, StartedAt: l.StartedAt, FinishedAt: l.FinishedAt, CreatedAt: l.CreatedAt,
	}
}

func ToNoteResponse(n *domain.Note) NoteResponse {
	return NoteResponse{
		ID: n.ID, BookID: n.BookID, VerseID: n.VerseID, Content: n.Content,
		CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}
}
