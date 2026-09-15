// Package pagination implements opaque cursor pagination, used instead of
// OFFSET/LIMIT so listing large tables (verses, books) stays O(1) regardless of depth.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Cursor is the decoded shape of the opaque "after" token. SortValue is the value of
// the column the result set is ordered by for the last row of the previous page
// (e.g. created_at as RFC3339, or a zero-padded numeric string); ID breaks ties.
type Cursor struct {
	SortValue string `json:"sv"`
	ID        int64  `json:"id"`
}

func Encode(sortValue string, id int64) string {
	raw, _ := json.Marshal(Cursor{SortValue: sortValue, ID: id})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func Decode(token string) (*Cursor, error) {
	if token == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	return &c, nil
}

// NormalizeLimit clamps a client-supplied page size into (0, MaxLimit], defaulting
// non-positive values to DefaultLimit.
func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// Page is the standard "meta" block returned alongside a list payload.
type Page struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}
