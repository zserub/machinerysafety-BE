package domain

// Person represents a user in the system.
type Person struct {
	PersonID    int16
	PersonName  *string
	PersonEmail *string
	Username    *string
	PwHash      *string
	RoleID      int16
	Locked      bool

	// Joined fields
	RoleName    string
	Permissions []Permission
}

// HasPermission checks if the person has a specific permission.
func (p *Person) HasPermission(resource, action string) bool {
	for _, perm := range p.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	return false
}

// Permission represents an RBAC permission.
type Permission struct {
	PermissionID int16
	Resource     string
	Action       string
}

// Role represents a user role.
type Role struct {
	RoleID   int16
	RoleName string
}
