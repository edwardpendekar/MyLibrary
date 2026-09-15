package dto

import "time"

type DashboardStatsResponse struct {
	TotalBooks       int64 `json:"total_books"`
	PublishedBooks   int64 `json:"published_books"`
	TotalChapters    int64 `json:"total_chapters"`
	TotalVerses      int64 `json:"total_verses"`
	TotalUsers       int64 `json:"total_users"`
	ImportsLast30Day int64 `json:"imports_last_30_day"`
}

type AuditLogResponse struct {
	ID         int64                  `json:"id"`
	UserID     *int64                 `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   *string                `json:"entity_id,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
