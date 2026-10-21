package metabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zserub/machinerysafety-BE/internal/config"
)

// Client provides Metabase API operations.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates a new Metabase client.
func NewClient(cfg *config.Config, logger *slog.Logger) *Client {
	logger = logger.With("component", "metabase.client")

	// Remove trailing slash from base URL
	baseURL := strings.TrimSuffix(cfg.MetabaseURL, "/")

	logger.Info("initializing Metabase client", "base_url", baseURL)

	return &Client{
		baseURL: baseURL,
		apiKey:  cfg.MetabaseAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Ping verifies Metabase connectivity.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Metabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metabase health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// GetDashboard fetches a dashboard by ID.
func (c *Client) GetDashboard(ctx context.Context, id int) (*Dashboard, error) {
	c.logger.Debug("fetching dashboard", "dashboard_id", id)

	start := time.Now()
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/dashboard/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var dashboard Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode dashboard: %w", err)
	}

	c.logger.Debug("dashboard fetched",
		"dashboard_id", id,
		"name", dashboard.Name,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &dashboard, nil
}

// GetCard fetches a card (question) by ID.
func (c *Client) GetCard(ctx context.Context, id int) (*Card, error) {
	c.logger.Debug("fetching card", "card_id", id)

	start := time.Now()
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/card/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
	}

	c.logger.Debug("card fetched",
		"card_id", id,
		"name", card.Name,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &card, nil
}

// CreateDashboardPublicLink enables public sharing for a dashboard and returns the UUID.
// If already shared, returns the existing UUID.
func (c *Client) CreateDashboardPublicLink(ctx context.Context, dashboardID int) (string, error) {
	c.logger.Debug("creating dashboard public link", "dashboard_id", dashboardID)

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/dashboard/%d/public_link", dashboardID), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return "", err
	}

	var result PublicLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode public link response: %w", err)
	}

	c.logger.Info("dashboard public link created", "dashboard_id", dashboardID, "uuid", result.UUID)
	return result.UUID, nil
}

// CreateCardPublicLink enables public sharing for a card and returns the UUID.
// If already shared, returns the existing UUID.
func (c *Client) CreateCardPublicLink(ctx context.Context, cardID int) (string, error) {
	c.logger.Debug("creating card public link", "card_id", cardID)

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/card/%d/public_link", cardID), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return "", err
	}

	var result PublicLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode public link response: %w", err)
	}

	c.logger.Info("card public link created", "card_id", cardID, "uuid", result.UUID)
	return result.UUID, nil
}

// GetPublicDashboardURL returns the public embed URL for a dashboard.
func (c *Client) GetPublicDashboardURL(uuid string) string {
	return fmt.Sprintf("%s/public/dashboard/%s", c.baseURL, uuid)
}

// GetPublicCardURL returns the public embed URL for a card.
func (c *Client) GetPublicCardURL(uuid string) string {
	return fmt.Sprintf("%s/public/question/%s", c.baseURL, uuid)
}

// ListDashboards returns all dashboards.
func (c *Client) ListDashboards(ctx context.Context) ([]DashboardSummary, error) {
	c.logger.Debug("listing dashboards")

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/dashboard", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var dashboards []DashboardSummary
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return nil, fmt.Errorf("failed to decode dashboards: %w", err)
	}

	c.logger.Debug("dashboards listed", "count", len(dashboards))
	return dashboards, nil
}

// ListCards returns all cards (excluding archived).
func (c *Client) ListCards(ctx context.Context) ([]CardSummary, error) {
	c.logger.Debug("listing cards")

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/card", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var allCards []CardSummary
	if err := json.NewDecoder(resp.Body).Decode(&allCards); err != nil {
		return nil, fmt.Errorf("failed to decode cards: %w", err)
	}

	// Filter out archived cards
	cards := make([]CardSummary, 0, len(allCards))
	for _, card := range allCards {
		if !card.Archived {
			cards = append(cards, card)
		}
	}

	c.logger.Debug("cards listed", "count", len(cards))
	return cards, nil
}

// FindDashboardByName finds a dashboard by exact name match.
func (c *Client) FindDashboardByName(ctx context.Context, name string) (*DashboardSummary, error) {
	dashboards, err := c.ListDashboards(ctx)
	if err != nil {
		return nil, err
	}

	for _, d := range dashboards {
		if d.Name == name {
			return &d, nil
		}
	}
	return nil, nil
}

// FindCardByName finds a card by exact name match.
func (c *Client) FindCardByName(ctx context.Context, name string) (*CardSummary, error) {
	cards, err := c.ListCards(ctx)
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		if card.Name == name {
			return &card, nil
		}
	}
	return nil, nil
}

// CreateCard creates a new Metabase card (question).
func (c *Client) CreateCard(ctx context.Context, req CreateCardRequest) (*CreateCardResponse, error) {
	c.logger.Debug("creating card", "name", req.Name)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal card request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/card", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result CreateCardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode card response: %w", err)
	}

	c.logger.Info("card created", "id", result.ID, "name", result.Name)
	return &result, nil
}

// CreateDashboard creates a new Metabase dashboard.
func (c *Client) CreateDashboard(ctx context.Context, req CreateDashboardRequest) (*CreateDashboardResponse, error) {
	c.logger.Debug("creating dashboard", "name", req.Name)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dashboard request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/dashboard", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result CreateDashboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode dashboard response: %w", err)
	}

	c.logger.Info("dashboard created", "id", result.ID, "name", result.Name)
	return &result, nil
}

// UpdateDashboardCards adds or updates cards on a dashboard.
func (c *Client) UpdateDashboardCards(ctx context.Context, dashboardID int, req UpdateDashboardCardsRequest) error {
	c.logger.Debug("updating dashboard cards", "dashboard_id", dashboardID, "card_count", len(req.Dashcards))

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal cards request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/dashboard/%d", dashboardID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	c.logger.Info("dashboard cards updated", "dashboard_id", dashboardID, "card_count", len(req.Dashcards))
	return nil
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// checkResponse checks the HTTP response for errors.
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusBadRequest:
		if strings.Contains(string(body), "public sharing") {
			return ErrPublicSharingOff
		}
		return &MetabaseError{
			StatusCode: resp.StatusCode,
			Message:    "bad request",
			Details:    string(body),
		}
	default:
		return &MetabaseError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Details:    string(body),
		}
	}
}
