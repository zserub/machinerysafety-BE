package metabase

import "fmt"

// Dashboard represents a Metabase dashboard.
type Dashboard struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicUUID string `json:"public_uuid"`
}

// Card represents a Metabase card (question).
type Card struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicUUID string `json:"public_uuid"`
}

// PublicLinkResponse is the response from creating a public link.
type PublicLinkResponse struct {
	UUID string `json:"uuid"`
}

// MetabaseError represents an error from the Metabase API.
type MetabaseError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *MetabaseError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("metabase error %d: %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("metabase error %d: %s", e.StatusCode, e.Message)
}

// Sentinel errors for common Metabase API errors.
var (
	ErrNotFound         = &MetabaseError{StatusCode: 404, Message: "resource not found"}
	ErrUnauthorized     = &MetabaseError{StatusCode: 401, Message: "unauthorized"}
	ErrPublicSharingOff = &MetabaseError{StatusCode: 400, Message: "public sharing is disabled"}
)

// ===== Card (Question) Creation Types =====

// CreateCardRequest represents the request body for creating a card (question).
type CreateCardRequest struct {
	Name                  string                 `json:"name"`
	DatasetQuery          DatasetQuery           `json:"dataset_query"`
	Display               string                 `json:"display"` // scalar, bar, line, pie, table, row, pivot
	VisualizationSettings map[string]any         `json:"visualization_settings"`
	CollectionID          *int                   `json:"collection_id,omitempty"`
	Description           *string                `json:"description,omitempty"`
}

// DatasetQuery defines the query for a card.
type DatasetQuery struct {
	Database int         `json:"database"`
	Type     string      `json:"type"` // "native" for SQL
	Native   NativeQuery `json:"native"`
}

// NativeQuery holds the raw SQL and template tags for parameters.
type NativeQuery struct {
	Query        string                 `json:"query"`
	TemplateTags map[string]TemplateTag `json:"template-tags,omitempty"`
}

// TemplateTag defines a parameter for native queries.
type TemplateTag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display-name"`
	Type        string `json:"type"` // "date", "dimension", "text", "number"
	Required    bool   `json:"required,omitempty"`
	Default     any    `json:"default,omitempty"`
	Dimension   any    `json:"dimension,omitempty"` // For dimension-type filters
}

// CreateCardResponse is the response from creating a card.
type CreateCardResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ===== Dashboard Creation Types =====

// CreateDashboardRequest represents the request body for creating a dashboard.
type CreateDashboardRequest struct {
	Name         string               `json:"name"`
	Description  *string              `json:"description,omitempty"`
	CollectionID *int                 `json:"collection_id,omitempty"`
	Parameters   []DashboardParameter `json:"parameters,omitempty"`
}

// DashboardParameter defines a filter parameter on the dashboard.
type DashboardParameter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Type      string `json:"type"` // "date/all-options", "date/range", etc.
	Default   any    `json:"default,omitempty"`
	SectionID string `json:"sectionId,omitempty"`
}

// CreateDashboardResponse is the response from creating a dashboard.
type CreateDashboardResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DashboardCardMapping maps a card to a position on the dashboard.
type DashboardCardMapping struct {
	ID                int                `json:"id,omitempty"` // Required for updates
	CardID            *int               `json:"card_id"`
	Row               int                `json:"row"`
	Col               int                `json:"col"`
	SizeX             int                `json:"size_x"`
	SizeY             int                `json:"size_y"`
	ParameterMappings []ParameterMapping `json:"parameter_mappings,omitempty"`
}

// ParameterMapping links a dashboard parameter to a card's template tag.
type ParameterMapping struct {
	ParameterID string `json:"parameter_id"`
	CardID      int    `json:"card_id"`
	Target      any    `json:"target"` // ["dimension", ["template-tag", "date_filter"]]
}

// UpdateDashboardCardsRequest represents the request to add/update cards on a dashboard.
type UpdateDashboardCardsRequest struct {
	Dashcards []DashboardCardMapping `json:"dashcards"`
}

// ===== List Response Types =====

// DashboardSummary is a minimal dashboard representation for listing.
type DashboardSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CardSummary is a minimal card representation for listing.
type CardSummary struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
}
