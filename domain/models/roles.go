package models

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

func ValidateRole(role string) bool {
	switch role {
	case RoleAdmin, RoleMember:
		return true
	}
	return false
}
