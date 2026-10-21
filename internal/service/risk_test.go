package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateRiskInput_Fields(t *testing.T) {
	deadline := time.Now().Add(24 * time.Hour)
	cost := 1000.50
	expenseTypeID := 1
	responsibleID := int16(5)
	linkedRiskID := int64(10)
	mocNumber := 12345
	categoryID := 2
	action := "Install guard"

	input := CreateRiskInput{
		RiskType:          1,
		LinkedRiskID:      &linkedRiskID,
		Machines:          "Machine A, Machine B",
		LineID:            3,
		CategoryID:        &categoryID,
		DangerDescription: "Pinch point hazard",
		RiskLevelID:       4,
		Action:            &action,
		ResponsibleID:     &responsibleID,
		MOCNumber:         &mocNumber,
		Deadline:          &deadline,
		Cost:              &cost,
		ExpenseTypeID:     &expenseTypeID,
	}

	assert.Equal(t, 1, input.RiskType)
	assert.Equal(t, int64(10), *input.LinkedRiskID)
	assert.Equal(t, "Machine A, Machine B", input.Machines)
	assert.Equal(t, int16(3), input.LineID)
	assert.Equal(t, 2, *input.CategoryID)
	assert.Equal(t, "Pinch point hazard", input.DangerDescription)
	assert.Equal(t, int16(4), input.RiskLevelID)
	assert.Equal(t, "Install guard", *input.Action)
	assert.Equal(t, int16(5), *input.ResponsibleID)
	assert.Equal(t, 12345, *input.MOCNumber)
	assert.NotNil(t, input.Deadline)
	assert.Equal(t, 1000.50, *input.Cost)
	assert.Equal(t, 1, *input.ExpenseTypeID)
}

func TestUpdateRiskInput_Fields(t *testing.T) {
	machines := "Updated Machine"
	lineID := int16(2)
	categoryID := 3
	dangerDesc := "Updated danger"
	riskLevelID := int16(5)
	action := "Updated action"
	responsibleID := int16(10)
	mocNumber := 54321
	deadline := time.Now().Add(48 * time.Hour)
	actionStateID := int16(2)
	implDate := time.Now()
	cost := 2000.75
	expenseTypeID := 2
	linkedRiskID := int64(20)
	statusReason := "Progressing well"
	deadlineReason := "Extended due to parts delay"

	input := UpdateRiskInput{
		Machines:             &machines,
		LineID:               &lineID,
		CategoryID:           &categoryID,
		DangerDescription:    &dangerDesc,
		RiskLevelID:          &riskLevelID,
		Action:               &action,
		ResponsibleID:        &responsibleID,
		MOCNumber:            &mocNumber,
		Deadline:             &deadline,
		ActionStateID:        &actionStateID,
		ImplementationDate:   &implDate,
		Cost:                 &cost,
		ExpenseTypeID:        &expenseTypeID,
		LinkedRiskID:         &linkedRiskID,
		StatusChangeReason:   &statusReason,
		DeadlineChangeReason: &deadlineReason,
	}

	assert.Equal(t, "Updated Machine", *input.Machines)
	assert.Equal(t, int16(2), *input.LineID)
	assert.Equal(t, 3, *input.CategoryID)
	assert.Equal(t, "Updated danger", *input.DangerDescription)
	assert.Equal(t, int16(5), *input.RiskLevelID)
	assert.Equal(t, "Updated action", *input.Action)
	assert.Equal(t, int16(10), *input.ResponsibleID)
	assert.Equal(t, 54321, *input.MOCNumber)
	assert.NotNil(t, input.Deadline)
	assert.Equal(t, int16(2), *input.ActionStateID)
	assert.NotNil(t, input.ImplementationDate)
	assert.Equal(t, 2000.75, *input.Cost)
	assert.Equal(t, 2, *input.ExpenseTypeID)
	assert.Equal(t, int64(20), *input.LinkedRiskID)
	assert.Equal(t, "Progressing well", *input.StatusChangeReason)
	assert.Equal(t, "Extended due to parts delay", *input.DeadlineChangeReason)
}

func TestUpdateRiskInput_PartialUpdate(t *testing.T) {
	// Test that partial updates work (only some fields set)
	machines := "Only machines updated"

	input := UpdateRiskInput{
		Machines: &machines,
	}

	assert.Equal(t, "Only machines updated", *input.Machines)
	assert.Nil(t, input.LineID)
	assert.Nil(t, input.CategoryID)
	assert.Nil(t, input.DangerDescription)
	assert.Nil(t, input.RiskLevelID)
	assert.Nil(t, input.Action)
	assert.Nil(t, input.ResponsibleID)
	assert.Nil(t, input.Deadline)
	assert.Nil(t, input.ActionStateID)
	assert.Nil(t, input.LinkedRiskID)
	assert.Nil(t, input.StatusChangeReason)
	assert.Nil(t, input.DeadlineChangeReason)
}
