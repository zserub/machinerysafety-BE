package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCommentInput_Fields(t *testing.T) {
	input := CreateCommentInput{
		RiskID:  100,
		Content: "This is a test comment",
	}

	assert.Equal(t, int64(100), input.RiskID)
	assert.Equal(t, "This is a test comment", input.Content)
}

func TestCreateCommentInput_EmptyContent(t *testing.T) {
	input := CreateCommentInput{
		RiskID:  100,
		Content: "",
	}

	assert.Equal(t, int64(100), input.RiskID)
	assert.Equal(t, "", input.Content)
}

func TestCreateCommentInput_LongContent(t *testing.T) {
	longContent := make([]byte, 500)
	for i := range longContent {
		longContent[i] = 'a'
	}

	input := CreateCommentInput{
		RiskID:  100,
		Content: string(longContent),
	}

	assert.Equal(t, int64(100), input.RiskID)
	assert.Len(t, input.Content, 500)
}
