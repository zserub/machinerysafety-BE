package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zserub/machinerysafety-BE/internal/utils"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := utils.HashPassword(password, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Hash should be different each time due to random salt
	hash2, err := utils.HashPassword(password, 10)
	assert.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestVerifyPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, _ := utils.HashPassword(password, 10)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "not-a-valid-bcrypt-hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.VerifyPassword(tt.password, tt.hash)
			assert.Equal(t, tt.want, got)
		})
	}
}
