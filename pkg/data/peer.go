package data

// Peer associates a peer with 0+ groups
type Peer struct {
	ID                     string   `yaml:"id" json:"id"`
	Name                   string   `yaml:"name" json:"name"`
	Groups                 []Group  `yaml:"-" json:"groups"`
	GroupNames             []string `yaml:"groups"`
	SSHEnabled             bool     `yaml:"ssh_enabled" json:"ssh_enabled"`
	ExpirationDisabled     bool     `yaml:"expiration_disabled"`
	LoginExpirationEnabled bool     `json:"login_expiration_enabled"`
	UserID                 string   `json:"user_id"`
}
