package repo

import (
	_ "embed"

	"github.com/midir99/sqload"
)

//go:embed sql/risk.sql
var riskSQL string

//go:embed sql/comment.sql
var commentSQL string

//go:embed sql/metadata.sql
var metadataSQL string

//go:embed sql/person.sql
var personSQL string

//go:embed sql/preference.sql
var preferenceSQL string

// RiskQueries contains all risk-related SQL queries.
type RiskQueries struct {
	CountRisks              string `query:"count_risks"`
	ListRisksBase           string `query:"list_risks_base"`
	GetByID                 string `query:"get_by_id"`
	GetPhotos               string `query:"get_photos"`
	GetStatusHistory        string `query:"get_status_history"`
	GetDeadlineHistory      string `query:"get_deadline_history"`
	GetLinkedRiskSummary    string `query:"get_linked_risk_summary"`
	Create                  string `query:"create"`
	Update                  string `query:"update"`
	Delete                  string `query:"delete"`
	DeleteStatusHistory     string `query:"delete_status_history"`
	DeleteDeadlineHistory   string `query:"delete_deadline_history"`
	DeleteComments          string `query:"delete_comments"`
	DeletePhotos            string `query:"delete_photos"`
	ClearLinkedRiskRefs     string `query:"clear_linked_risk_references"`
	GetRiskType             string `query:"get_risk_type"`
	GetCurrentStatus        string `query:"get_current_status"`
	GetCurrentDeadline      string `query:"get_current_deadline"`
	InsertStatusHistory     string `query:"insert_status_history"`
	InsertDeadlineHistory   string `query:"insert_deadline_history"`
}

// CommentQueries contains all comment-related SQL queries.
type CommentQueries struct {
	ListByRiskID string `query:"list_by_risk_id"`
	Create       string `query:"create"`
	GetByID      string `query:"get_by_id"`
}

// MetadataQueries contains all metadata-related SQL queries.
type MetadataQueries struct {
	GetRiskLevels              string `query:"get_risk_levels"`
	GetStatuses                string `query:"get_statuses"`
	GetCategories              string `query:"get_categories"`
	GetExpenseTypes            string `query:"get_expense_types"`
	GetLines                   string `query:"get_lines"`
	GetResponsibles            string `query:"get_responsibles"`
	GetLineIDByName            string `query:"get_line_id_by_name"`
	GetCategoryIDByName        string `query:"get_category_id_by_name"`
	GetRiskLevelIDByName       string `query:"get_risk_level_id_by_name"`
	GetExpenseTypeIDByName     string `query:"get_expense_type_id_by_name"`
	GetResponsibleIDByUsername string `query:"get_responsible_id_by_username"`
	GetResponsibleIDByName     string `query:"get_responsible_id_by_name"`
	GetStatusIDByName          string `query:"get_status_id_by_name"`
}

// PersonQueries contains all person-related SQL queries.
type PersonQueries struct {
	GetByID            string `query:"get_by_id"`
	GetByUsername      string `query:"get_by_username"`
	GetPermissions     string `query:"get_permissions"`
	CreateInitialAdmin string `query:"create_initial_admin"`
}

// PreferenceQueries contains all preference-related SQL queries.
type PreferenceQueries struct {
	GetByPersonID     string `query:"get_by_person_id"`
	Update            string `query:"update"`
	UpdateWithColumns string `query:"update_with_columns"`
	CreateDefault     string `query:"create_default"`
}

// Loaded queries (initialized at startup)
var (
	riskQ       = sqload.MustLoadFromString[RiskQueries](riskSQL)
	commentQ    = sqload.MustLoadFromString[CommentQueries](commentSQL)
	metadataQ   = sqload.MustLoadFromString[MetadataQueries](metadataSQL)
	personQ     = sqload.MustLoadFromString[PersonQueries](personSQL)
	preferenceQ = sqload.MustLoadFromString[PreferenceQueries](preferenceSQL)
)
