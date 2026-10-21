package domain

import "time"

// Comment represents a comment on a risk.
type Comment struct {
	ID         int64
	RiskID     int64
	AuthorID   int16
	Content    string
	CreatedAt  time.Time
	EditedAt   *time.Time
	AuthorName string // Joined
}
