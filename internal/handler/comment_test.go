package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

func TestCreateCommentRequest_Validation(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantValid bool
	}{
		{
			name:      "valid request",
			json:      `{"content":"This is a comment"}`,
			wantValid: true,
		},
		{
			name:      "empty content",
			json:      `{"content":""}`,
			wantValid: true, // JSON parsing succeeds, validation happens in handler
		},
		{
			name:      "long content",
			json:      `{"content":"` + strings.Repeat("a", 500) + `"}`,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req CreateCommentRequest
			err := json.Unmarshal([]byte(tt.json), &req)

			if tt.wantValid {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentHandler_CreateValidation(t *testing.T) {
	tests := []struct {
		name           string
		riskId         string
		person         *domain.Person
		requestBody    map[string]interface{}
		wantStatusCode int
	}{
		{
			name:           "unauthorized - no person",
			riskId:         "1",
			person:         nil,
			requestBody:    map[string]interface{}{"content": "test"},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:   "invalid riskId - letters",
			riskId: "abc",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"content": "test comment",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "missing content",
			riskId: "1",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"content": "",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "content too long",
			riskId: "1",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"content": strings.Repeat("a", 501),
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &CommentHandler{
				commentService: nil,
				logger:         nil,
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/entries/"+tt.riskId+"/comments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("riskId", tt.riskId)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

			if tt.person != nil {
				ctx = context.WithValue(ctx, ContextKeyPerson, tt.person)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			h.Create(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestCommentHandler_ListIdParsing(t *testing.T) {
	tests := []struct {
		name           string
		riskId         string
		wantStatusCode int
	}{
		{
			name:           "invalid riskId - letters",
			riskId:         "abc",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid riskId - special chars",
			riskId:         "1-2-3",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid riskId - empty",
			riskId:         "",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &CommentHandler{
				commentService: nil,
				logger:         nil,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/entries/"+tt.riskId+"/comments", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("riskId", tt.riskId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			h.List(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestToCommentResponse(t *testing.T) {
	createdAt := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)
	editedAt := time.Date(2026, 2, 14, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		comment *domain.Comment
		want    CommentResponse
	}{
		{
			name: "full comment",
			comment: &domain.Comment{
				ID:         1,
				RiskID:     100,
				AuthorID:   5,
				AuthorName: "John Smith",
				Content:    "Test comment",
				CreatedAt:  createdAt,
				EditedAt:   &editedAt,
			},
			want: CommentResponse{
				ID:         1,
				RiskID:     100,
				AuthorID:   5,
				AuthorName: "John Smith",
				Content:    "Test comment",
				CreatedAt:  "2026-02-14T10:30:00Z",
				EditedAt:   stringPtr("2026-02-14T11:00:00Z"),
			},
		},
		{
			name: "comment without edit",
			comment: &domain.Comment{
				ID:         2,
				RiskID:     100,
				AuthorID:   5,
				AuthorName: "Jane Doe",
				Content:    "Another comment",
				CreatedAt:  createdAt,
				EditedAt:   nil,
			},
			want: CommentResponse{
				ID:         2,
				RiskID:     100,
				AuthorID:   5,
				AuthorName: "Jane Doe",
				Content:    "Another comment",
				CreatedAt:  "2026-02-14T10:30:00Z",
				EditedAt:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCommentResponse(tt.comment)
			assert.Equal(t, tt.want.ID, result.ID)
			assert.Equal(t, tt.want.RiskID, result.RiskID)
			assert.Equal(t, tt.want.AuthorID, result.AuthorID)
			assert.Equal(t, tt.want.AuthorName, result.AuthorName)
			assert.Equal(t, tt.want.Content, result.Content)
			assert.Equal(t, tt.want.CreatedAt, result.CreatedAt)
			if tt.want.EditedAt == nil {
				assert.Nil(t, result.EditedAt)
			} else {
				assert.Equal(t, *tt.want.EditedAt, *result.EditedAt)
			}
		})
	}
}

func TestToCommentResponses(t *testing.T) {
	createdAt := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)

	comments := []domain.Comment{
		{
			ID:         1,
			RiskID:     100,
			AuthorID:   5,
			AuthorName: "John Smith",
			Content:    "First comment",
			CreatedAt:  createdAt,
		},
		{
			ID:         2,
			RiskID:     100,
			AuthorID:   6,
			AuthorName: "Jane Doe",
			Content:    "Second comment",
			CreatedAt:  createdAt.Add(time.Hour),
		},
	}

	result := toCommentResponses(comments)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(2), result[1].ID)
}

func TestToCommentResponses_Empty(t *testing.T) {
	comments := []domain.Comment{}
	result := toCommentResponses(comments)
	assert.Len(t, result, 0)
	assert.NotNil(t, result)
}

func stringPtr(s string) *string {
	return &s
}
