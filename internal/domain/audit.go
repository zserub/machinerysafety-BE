package domain

import "time"

// StatusHistoryItem records a status change for a risk.
type StatusHistoryItem struct {
	ID            int64
	RiskID        int64
	OldStatus     int16
	NewStatus     int16
	Reason        *string
	ChangedBy     int16
	ChangedAt     time.Time
	OldStateName  string
	NewStateName  string
	ChangedByName string
}

// DeadlineHistoryItem records a deadline change for a risk.
type DeadlineHistoryItem struct {
	ID            int64
	RiskID        int64
	OldDeadline   time.Time
	NewDeadline   time.Time
	Reason        string
	ChangedBy     int16
	ChangedAt     time.Time
	ChangedByName string
}

// AuditLogEntry represents a generic audit log record.
type AuditLogEntry struct {
	ID        int64
	TableName string
	RecordID  int64
	Action    string // INSERT, UPDATE, DELETE
	OldValues any    // jsonb
	NewValues any    // jsonb
	ChangedBy int16
	ChangedAt time.Time
}
