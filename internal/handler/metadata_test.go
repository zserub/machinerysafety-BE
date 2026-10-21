package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

func TestMetadata_JSONSerialization(t *testing.T) {
	metadata := &domain.Metadata{
		RiskLevels: []domain.RiskLevelOption{
			{Value: "Critical", Order: 1},
			{Value: "High", Order: 2},
			{Value: "Medium", Order: 3},
		},
		Statuses: []domain.StatusOption{
			{Value: "OPEN", Order: 1, IsTerminal: false},
			{Value: "ASSIGNED", Order: 2, IsTerminal: false},
			{Value: "IN_PROGRESS", Order: 3, IsTerminal: false},
			{Value: "CLOSED", Order: 4, IsTerminal: true},
		},
		Categories: []domain.CategoryOption{
			{Value: "Mechanical", Order: 1},
			{Value: "Electrical", Order: 2},
		},
		ExpenseTypes: []domain.ExpenseTypeOption{
			{Value: "CAPEX", Order: 1},
			{Value: "OPEX", Order: 2},
		},
		Lines: []domain.LineOption{
			{Value: "Line A", Order: 1},
			{Value: "Line B", Order: 2},
		},
		Responsibles: []domain.ResponsibleOption{
			{Value: "jdoe", Name: "John Doe", Order: 1},
			{Value: "jsmith", Name: "Jane Smith", Order: 2},
		},
	}

	data, err := json.Marshal(metadata)
	assert.NoError(t, err)

	var result domain.Metadata
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(result.RiskLevels))
	assert.Equal(t, 4, len(result.Statuses))
	assert.Equal(t, 2, len(result.Categories))
	assert.Equal(t, 2, len(result.ExpenseTypes))
	assert.Equal(t, 2, len(result.Lines))
	assert.Equal(t, 2, len(result.Responsibles))

	assert.Equal(t, "Critical", result.RiskLevels[0].Value)
	assert.Equal(t, true, result.Statuses[3].IsTerminal)
	assert.Equal(t, "John Doe", result.Responsibles[0].Name)
}

func TestMetadata_JSONFieldNames(t *testing.T) {
	metadata := &domain.Metadata{
		RiskLevels:   []domain.RiskLevelOption{{Value: "Test", Order: 1}},
		Statuses:     []domain.StatusOption{{Value: "OPEN", Order: 1, IsTerminal: false}},
		Categories:   []domain.CategoryOption{{Value: "Cat", Order: 1}},
		ExpenseTypes: []domain.ExpenseTypeOption{{Value: "CAPEX", Order: 1}},
		Lines:        []domain.LineOption{{Value: "Line", Order: 1}},
		Responsibles: []domain.ResponsibleOption{{Value: "user", Name: "User", Order: 1}},
	}

	data, err := json.Marshal(metadata)
	assert.NoError(t, err)

	jsonStr := string(data)
	// Verify JSON field names are camelCase
	assert.Contains(t, jsonStr, `"riskLevels"`)
	assert.Contains(t, jsonStr, `"statuses"`)
	assert.Contains(t, jsonStr, `"categories"`)
	assert.Contains(t, jsonStr, `"expenseTypes"`)
	assert.Contains(t, jsonStr, `"lines"`)
	assert.Contains(t, jsonStr, `"responsibles"`)
	assert.Contains(t, jsonStr, `"isTerminal"`)
}

func TestRiskLevelOption_Structure(t *testing.T) {
	opt := domain.RiskLevelOption{
		Value: "High",
		Order: 2,
	}

	data, err := json.Marshal(opt)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "High", result["value"])
	assert.Equal(t, float64(2), result["order"])
}

func TestStatusOption_Structure(t *testing.T) {
	opt := domain.StatusOption{
		Value:      "CLOSED",
		Order:      4,
		IsTerminal: true,
	}

	data, err := json.Marshal(opt)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "CLOSED", result["value"])
	assert.Equal(t, float64(4), result["order"])
	assert.Equal(t, true, result["isTerminal"])
}

func TestResponsibleOption_Structure(t *testing.T) {
	opt := domain.ResponsibleOption{
		Value: "jdoe",
		Name:  "John Doe",
		Order: 1,
	}

	data, err := json.Marshal(opt)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "jdoe", result["value"])
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, float64(1), result["order"])
}
