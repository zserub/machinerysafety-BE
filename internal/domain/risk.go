package domain

import "time"

// Risk types
const (
	RiskTypeShortTerm = 1
	RiskTypeLongTerm  = 2
)

// Risk represents a safety risk entry.
type Risk struct {
	ID                 int64
	RiskType           int // 1=short-term, 2=long-term
	LinkedRiskID       *int64
	CreationDate       time.Time
	Machines           string
	LineID             int16
	CategoryID         *int
	DangerDescription  string
	RiskLevelID        int16
	Action             *string
	ResponsibleID      *int16
	MOCNumber          *int
	Deadline           *time.Time
	ActionStateID      int16
	ImplementationDate *time.Time
	EntryModifiedAt    time.Time
	CreatedByID        int16
	CommittedAt        *time.Time
	Cost               *float64
	ExpenseTypeID      *int

	// Joined/computed fields (not always populated)
	LineName        string
	CategoryName    *string
	RiskLevelName   string
	StateName       string
	ResponsibleName *string
	CreatedByName   *string
	ExpenseTypeName *string
	LinkedRisk      *LinkedRiskSummary
	Photos          []Photo
	StatusHistory   []StatusHistoryItem
	DeadlineHistory []DeadlineHistoryItem
}

// LinkedRiskSummary provides a minimal view of a linked risk.
type LinkedRiskSummary struct {
	ID                int64
	Machines          string
	LineName          string
	DangerDescription string
	RiskType          int
}
