package domain

// Category represents a risk category.
type Category struct {
	ID           int
	CategoryName string
}

// ExpenseType represents CAPEX or OPEX.
type ExpenseType struct {
	ID   int
	Type string // CAPEX, OPEX
}

// Line represents a production line.
type Line struct {
	LineID   int16
	LineName string
}

// RiskLevel represents a risk severity level.
type RiskLevel struct {
	RiskLevelID int16
	RiskLevel   string
	Severity    int // 1-5 for sorting
}

// ActionState represents the state of a risk action.
type ActionState struct {
	StateID    int16
	State      string // OPEN, ASSIGNED, IN_PROGRESS, CLOSED
	IsTerminal bool
}

// Metadata holds all reference data for the frontend.
type Metadata struct {
	RiskLevels   []RiskLevelOption   `json:"riskLevels"`
	Statuses     []StatusOption      `json:"statuses"`
	Categories   []CategoryOption    `json:"categories"`
	ExpenseTypes []ExpenseTypeOption `json:"expenseTypes"`
	Lines        []LineOption        `json:"lines"`
	Responsibles []ResponsibleOption `json:"responsibles"`
}

// RiskLevelOption is the API response format for risk levels.
type RiskLevelOption struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// StatusOption is the API response format for statuses.
type StatusOption struct {
	Value      string `json:"value"`
	Order      int    `json:"order"`
	IsTerminal bool   `json:"isTerminal"`
}

// CategoryOption is the API response format for categories.
type CategoryOption struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// ExpenseTypeOption is the API response format for expense types.
type ExpenseTypeOption struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// LineOption is the API response format for lines.
type LineOption struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

// ResponsibleOption is the API response format for responsible persons.
type ResponsibleOption struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}
