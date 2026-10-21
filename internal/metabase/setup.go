package metabase

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	// DashboardName is the name of the analytics dashboard.
	DashboardName = "Machinery Safety Analytics"
	// DateFilterID is the ID of the date filter parameter.
	DateFilterID = "date_filter"
)

// SetupService handles Metabase dashboard setup.
type SetupService struct {
	client     *Client
	databaseID int
	logger     *slog.Logger
}

// NewSetupService creates a new SetupService.
func NewSetupService(client *Client, databaseID int, logger *slog.Logger) *SetupService {
	return &SetupService{
		client:     client,
		databaseID: databaseID,
		logger:     logger.With("component", "metabase.setup"),
	}
}

// SetupResult holds the results of the setup operation.
type SetupResult struct {
	DashboardID   int
	DashboardName string
	CardsCreated  int
	CardsSkipped  int
	Skipped       bool // true if dashboard already existed
}

// Run executes the dashboard setup (idempotent).
func (s *SetupService) Run(ctx context.Context) (*SetupResult, error) {
	s.logger.Info("starting Metabase dashboard setup")

	// Check if dashboard already exists
	existing, err := s.client.FindDashboardByName(ctx, DashboardName)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing dashboards: %w", err)
	}
	if existing != nil {
		s.logger.Info("dashboard already exists, skipping setup", "id", existing.ID)
		return &SetupResult{
			DashboardID:   existing.ID,
			DashboardName: existing.Name,
			Skipped:       true,
		}, nil
	}

	// Create cards (idempotent by name)
	cardDefinitions := s.getCardDefinitions()
	cardIDs := make(map[string]int) // card name -> card ID
	created, skipped := 0, 0

	for _, def := range cardDefinitions {
		existingCard, err := s.client.FindCardByName(ctx, def.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing card %q: %w", def.Name, err)
		}
		if existingCard != nil {
			s.logger.Debug("card already exists, skipping", "name", def.Name, "id", existingCard.ID)
			cardIDs[def.Name] = existingCard.ID
			skipped++
			continue
		}

		card, err := s.client.CreateCard(ctx, def)
		if err != nil {
			return nil, fmt.Errorf("failed to create card %q: %w", def.Name, err)
		}
		cardIDs[def.Name] = card.ID
		created++
	}

	// Create dashboard with date filter parameter
	dashboard, err := s.client.CreateDashboard(ctx, CreateDashboardRequest{
		Name: DashboardName,
		Parameters: []DashboardParameter{{
			ID:   DateFilterID,
			Name: "Date Range",
			Slug: "date_range",
			Type: "date/all-options",
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	// Add cards to dashboard with layout
	cardMappings := s.getCardMappings(cardIDs)
	if err := s.client.UpdateDashboardCards(ctx, dashboard.ID, UpdateDashboardCardsRequest{
		Dashcards: cardMappings,
	}); err != nil {
		return nil, fmt.Errorf("failed to add cards to dashboard: %w", err)
	}

	s.logger.Info("dashboard setup completed",
		"dashboard_id", dashboard.ID,
		"cards_created", created,
		"cards_skipped", skipped,
	)

	return &SetupResult{
		DashboardID:   dashboard.ID,
		DashboardName: dashboard.Name,
		CardsCreated:  created,
		CardsSkipped:  skipped,
	}, nil
}

// getCardDefinitions returns all card definitions for the dashboard.
func (s *SetupService) getCardDefinitions() []CreateCardRequest {
	// Helper for scalar cards
	scalar := func(name, query string, vizSettings map[string]any) CreateCardRequest {
		if vizSettings == nil {
			vizSettings = map[string]any{}
		}
		return CreateCardRequest{
			Name:    name,
			Display: "scalar",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: query,
				},
			},
			VisualizationSettings: vizSettings,
		}
	}

	return []CreateCardRequest{
		// ===== Row 0: KPI Cards =====
		scalar("Total Risks", `
SELECT COUNT(*) AS total
FROM eng.risks r
`, nil),

		scalar("Open Risks", `
SELECT COUNT(*) AS open_risks
FROM eng.risks r
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
WHERE ast.is_terminal = FALSE
`, nil),

		scalar("Resolution Rate", `
SELECT
    ROUND(
        COUNT(*) FILTER (WHERE ast.is_terminal = TRUE) * 100.0 / NULLIF(COUNT(*), 0),
        1
    ) AS resolution_rate
FROM eng.risks r
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
`, map[string]any{
			"scalar.suffix": "%",
		}),

		scalar("Avg Resolution Time", `
SELECT
    ROUND(AVG(r.implementation_date - r.creation_date::date)::numeric, 1) AS avg_days
FROM eng.risks r
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
WHERE ast.is_terminal = TRUE
  AND r.implementation_date IS NOT NULL
`, map[string]any{
			"scalar.suffix": " days",
		}),

		// ===== Row 4: Severity Counts =====
		scalar("High Severity Risks", `
SELECT COUNT(*) AS count
FROM eng.risks r
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
WHERE rl.severity = 3
`, map[string]any{
			"scalar.field": "count",
		}),

		scalar("Medium Severity Risks", `
SELECT COUNT(*) AS count
FROM eng.risks r
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
WHERE rl.severity = 2
`, map[string]any{
			"scalar.field": "count",
		}),

		scalar("Low Severity Risks", `
SELECT COUNT(*) AS count
FROM eng.risks r
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
WHERE rl.severity = 1
`, map[string]any{
			"scalar.field": "count",
		}),

		scalar("Overdue Risks", `
SELECT COUNT(*) AS count
FROM eng.risks r
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
WHERE r.deadline < CURRENT_DATE
  AND ast.is_terminal = FALSE
`, map[string]any{
			"scalar.field": "count",
		}),

		scalar("Uncommitted Risks", `
SELECT COUNT(*) AS count
FROM eng.risks r
WHERE r.action_state_id = 1
`, map[string]any{
			"scalar.field": "count",
		}),

		// ===== Row 7: Charts =====
		{
			Name:    "Status Distribution",
			Display: "bar",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: `
SELECT
    ast.state AS status,
    COUNT(*) AS count
FROM eng.risks r
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
GROUP BY ast.state_id, ast.state
ORDER BY ast.state_id`,
				},
			},
			VisualizationSettings: map[string]any{
				"graph.dimensions":        []string{"status"},
				"graph.metrics":           []string{"count"},
				"graph.x_axis.title_text": "Status",
				"graph.y_axis.title_text": "Count",
			},
		},

		{
			Name:    "Risks by Category",
			Display: "pie",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: `
SELECT
    COALESCE(c.category_name, 'Uncategorized') AS category,
    COUNT(*) AS count
FROM eng.risks r
LEFT JOIN eng.categories c ON r.category_id = c.id
GROUP BY c.category_name
ORDER BY count DESC`,
				},
			},
			VisualizationSettings: map[string]any{
				"pie.dimension": "category",
				"pie.metric":    "count",
			},
		},

		// ===== Row 13: Tables (Matrix format) =====
		{
			Name:    "Risk Matrix - Lines",
			Display: "table",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: `
SELECT
    l.line_name AS "Line",
    COUNT(*) FILTER (WHERE rl.severity = 3) AS "High",
    COUNT(*) FILTER (WHERE rl.severity = 2) AS "Medium",
    COUNT(*) FILTER (WHERE rl.severity = 1) AS "Low",
    COUNT(*) AS "Total"
FROM eng.risks r
JOIN eng.lines l ON r.line_id = l.line_id
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
GROUP BY l.line_name
ORDER BY l.line_name`,
				},
			},
			VisualizationSettings: map[string]any{},
		},

		{
			Name:    "Risk Matrix - Responsible",
			Display: "table",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: `
SELECT
    COALESCE(p.person_name, 'Unassigned') AS "Responsible",
    COUNT(*) FILTER (WHERE rl.severity = 3) AS "High",
    COUNT(*) FILTER (WHERE rl.severity = 2) AS "Medium",
    COUNT(*) FILTER (WHERE rl.severity = 1) AS "Low",
    COUNT(*) AS "Total"
FROM eng.risks r
LEFT JOIN eng.persons p ON r.responsible_id = p.person_id
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
GROUP BY p.person_name
ORDER BY "Total" DESC`,
				},
			},
			VisualizationSettings: map[string]any{},
		},

		// ===== Row 20: Burndown Chart =====
		{
			Name:    "Monthly Resolution Burndown",
			Display: "line",
			DatasetQuery: DatasetQuery{
				Database: s.databaseID,
				Type:     "native",
				Native: NativeQuery{
					Query: `
SELECT
    DATE_TRUNC('month', r.implementation_date) AS month,
    rl.risk_level AS level,
    COUNT(*) AS resolved
FROM eng.risks r
JOIN eng.risk_levels rl ON r.risk_level_id = rl.risk_level_id
JOIN eng.action_states ast ON r.action_state_id = ast.state_id
WHERE ast.is_terminal = TRUE
  AND r.implementation_date IS NOT NULL
GROUP BY DATE_TRUNC('month', r.implementation_date), rl.risk_level, rl.severity
ORDER BY month, rl.severity`,
				},
			},
			VisualizationSettings: map[string]any{
				"graph.dimensions":        []string{"month", "level"},
				"graph.metrics":           []string{"resolved"},
				"graph.x_axis.title_text": "Month",
				"graph.y_axis.title_text": "Resolved Risks",
			},
		},
	}
}

// getCardMappings returns the dashboard card positions.
func (s *SetupService) getCardMappings(cardIDs map[string]int) []DashboardCardMapping {
	// Counter for generating unique negative IDs for new dashcards
	dashcardID := 0

	// Helper to create card mapping (without parameter mappings for now - can be linked in UI)
	card := func(name string, row, col, sizeX, sizeY int) DashboardCardMapping {
		dashcardID--
		cardID := cardIDs[name]
		return DashboardCardMapping{
			ID:     dashcardID, // Negative ID for new dashcards
			CardID: &cardID,
			Row:    row,
			Col:    col,
			SizeX:  sizeX,
			SizeY:  sizeY,
		}
	}

	return []DashboardCardMapping{
		// Row 0: KPIs (4 cards across)
		card("Total Risks", 0, 0, 6, 4),
		card("Open Risks", 0, 6, 6, 4),
		card("Resolution Rate", 0, 12, 6, 4),
		card("Avg Resolution Time", 0, 18, 6, 4),

		// Row 4: Severity Counts (5 cards)
		card("High Severity Risks", 4, 0, 5, 3),
		card("Medium Severity Risks", 4, 5, 5, 3),
		card("Low Severity Risks", 4, 10, 4, 3),
		card("Overdue Risks", 4, 14, 5, 3),
		card("Uncommitted Risks", 4, 19, 5, 3),

		// Row 7: Charts (2 cards)
		card("Status Distribution", 7, 0, 12, 6),
		card("Risks by Category", 7, 12, 12, 6),

		// Row 13: Tables (2 cards)
		card("Risk Matrix - Lines", 13, 0, 12, 7),
		card("Risk Matrix - Responsible", 13, 12, 12, 7),

		// Row 20: Burndown (1 card, full width)
		card("Monthly Resolution Burndown", 20, 0, 24, 7),
	}
}
