package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zserub/machinerysafety-BE/internal/domain"
)

func TestPerson_HasPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []domain.Permission
		resource    string
		action      string
		want        bool
	}{
		{
			name: "has exact permission",
			permissions: []domain.Permission{
				{Resource: "risks", Action: "read"},
				{Resource: "risks", Action: "create"},
			},
			resource: "risks",
			action:   "read",
			want:     true,
		},
		{
			name: "does not have permission",
			permissions: []domain.Permission{
				{Resource: "risks", Action: "read"},
			},
			resource: "risks",
			action:   "delete",
			want:     false,
		},
		{
			name:        "no permissions",
			permissions: nil,
			resource:    "risks",
			action:      "read",
			want:        false,
		},
		{
			name: "different resource",
			permissions: []domain.Permission{
				{Resource: "comments", Action: "read"},
			},
			resource: "risks",
			action:   "read",
			want:     false,
		},
		{
			name:        "empty permissions slice",
			permissions: []domain.Permission{},
			resource:    "risks",
			action:      "read",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person := &domain.Person{
				PersonID:    1,
				Permissions: tt.permissions,
			}

			got := person.HasPermission(tt.resource, tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}
