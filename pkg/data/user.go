package data

// User NetBird User to groups mapping
type User struct {
	Email       string   `yaml:"email" json:"email"`
	Groups      []string `yaml:"groups" json:"auto_groups"`
	ID          string   `json:"id"`
	Role        string   `yaml:"role" json:"role"`
	Blocked     bool     `json:"is_blocked"`
	ServiceUser bool     `json:"is_service_user"`
}

// GetRole returns role if valid, user otherwise
func (u User) GetRole() string {
	if u.Role != "admin" && u.Role != "user" && u.Role != "owner" {
		return "user"
	}

	return u.Role
}
