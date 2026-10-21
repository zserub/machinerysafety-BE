package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

func TestCreateRiskRequest_Validation(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantValid bool
	}{
		{
			name:      "valid request",
			json:      `{"riskType":1,"machines":"Machine A","lineId":"production-line-a","categoryId":"mechanical","dangerDescription":"Test","riskLevelId":"high","action":"Fix it"}`,
			wantValid: true,
		},
		{
			name:      "riskType 2 is valid",
			json:      `{"riskType":2,"machines":"Machine A","lineId":"production-line-a","categoryId":"mechanical","dangerDescription":"Test","riskLevelId":"high","action":"Fix it"}`,
			wantValid: true,
		},
		{
			name:      "with optional fields",
			json:      `{"riskType":1,"machines":"Machine A","lineId":"production-line-a","categoryId":"mechanical","dangerDescription":"Test","riskLevelId":"high","action":"Fix it","responsibleId":"john.smith","mocNumber":"12345","deadline":"2025-12-31","cost":1500.50,"expenseTypeId":"capex"}`,
			wantValid: true,
		},
		{
			name:      "with linked risk",
			json:      `{"riskType":1,"linkedRiskId":50,"machines":"Machine A","lineId":"production-line-a","categoryId":"mechanical","dangerDescription":"Test","riskLevelId":"high","action":"Fix it"}`,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req CreateRiskRequest
			err := json.Unmarshal([]byte(tt.json), &req)

			if tt.wantValid {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateRiskRequest_Validation(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "partial update - only machines",
			json: `{"machines":"Updated Machine"}`,
		},
		{
			name: "partial update - only status",
			json: `{"actionStateId":"in-progress","statusChangeReason":"Moving to assigned"}`,
		},
		{
			name: "partial update - deadline with reason",
			json: `{"deadline":"2025-06-01","deadlineChangeReason":"Extended for parts"}`,
		},
		{
			name: "full update",
			json: `{"machines":"M1","lineId":"production-line-b","categoryId":"electrical","dangerDescription":"Updated","riskLevelId":"medium","action":"New action","responsibleId":"john.smith","mocNumber":"99999","deadline":"2025-12-31","actionStateId":"completed","implementationDate":"2025-01-15","cost":5000,"expenseTypeId":"opex","linkedRiskId":100,"statusChangeReason":"Progress","deadlineChangeReason":"Delay"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateRiskRequest
			err := json.Unmarshal([]byte(tt.json), &req)
			assert.NoError(t, err)
		})
	}
}

func TestRiskHandler_CreateValidation(t *testing.T) {
	tests := []struct {
		name           string
		person         *domain.Person
		requestBody    map[string]interface{}
		wantStatusCode int
	}{
		{
			name:           "unauthorized - no person",
			person:         nil,
			requestBody:    map[string]interface{}{},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:   "invalid risk type 0",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"riskType":          0,
				"machines":          "M1",
				"dangerDescription": "Test",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "invalid risk type 3",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"riskType":          3,
				"machines":          "M1",
				"dangerDescription": "Test",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "missing machines",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"riskType":          1,
				"dangerDescription": "Test",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "missing danger description",
			person: &domain.Person{PersonID: 1},
			requestBody: map[string]interface{}{
				"riskType": 1,
				"machines": "M1",
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We're testing just the validation logic, not the full handler
			// So we create a handler with nil service (validation happens before service call)
			h := &RiskHandler{
				riskService: nil, // Validation happens before service call
				logger:      nil,
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/entries", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.person != nil {
				ctx := context.WithValue(req.Context(), ContextKeyPerson, tt.person)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// This will panic at service call if validation passes, which we expect for valid requests
			// For invalid requests, it should return before calling service
			defer func() {
				if r := recover(); r != nil {
					// Panic means validation passed but service was nil
					// This shouldn't happen for our invalid test cases
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			h.Create(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestRiskHandler_UpdateIdParsing(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		person         *domain.Person
		wantStatusCode int
	}{
		{
			name:           "invalid id - letters",
			id:             "abc",
			person:         &domain.Person{PersonID: 1},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid id - empty",
			id:             "",
			person:         &domain.Person{PersonID: 1},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no person",
			id:             "1",
			person:         nil,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &RiskHandler{
				riskService: nil,
				logger:      nil,
			}

			body, _ := json.Marshal(map[string]interface{}{"machines": "M1"})
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/entries/"+tt.id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

			if tt.person != nil {
				ctx = context.WithValue(ctx, ContextKeyPerson, tt.person)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.Update(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestRiskHandler_GetByIdParsing(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		wantStatusCode int
	}{
		{
			name:           "invalid id - letters",
			id:             "abc",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid id - special chars",
			id:             "1-2-3",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &RiskHandler{
				riskService: nil,
				logger:      nil,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/entries/"+tt.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			h.GetByID(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestRiskHandler_DeleteIdParsing(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		person         *domain.Person
		wantStatusCode int
	}{
		{
			name:           "invalid id",
			id:             "abc",
			person:         &domain.Person{PersonID: 1},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no person",
			id:             "1",
			person:         nil,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &RiskHandler{
				riskService: nil,
				logger:      nil,
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/entries/"+tt.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

			if tt.person != nil {
				ctx = context.WithValue(ctx, ContextKeyPerson, tt.person)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.Delete(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		def    int
		expect int
	}{
		{"empty string returns default", "", 10, 10},
		{"valid number", "5", 10, 5},
		{"invalid string returns default", "abc", 10, 10},
		{"zero is valid", "0", 10, 0},
		{"negative is valid", "-5", 10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntOrDefault(tt.input, tt.def)
			assert.Equal(t, tt.expect, result)
		})
	}
}
