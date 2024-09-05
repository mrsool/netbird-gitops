package data

// Config holds program configuration
type Config struct {
	AutoSync             string `yaml:"autoSync"`
	IndividualPeerGroups bool   `yaml:"individualPeerGroups"`
}

// CombinedConfig combined config of all files
type CombinedConfig struct {
	Config        Config         `yaml:"config"`
	Nameservers   []Nameserver   `yaml:"nameservers"`
	DNS           DNS            `yaml:"dns"`
	Peers         []Peer         `yaml:"peers"`
	Policies      []Policy       `yaml:"policies"`
	PostureChecks []PostureCheck `yaml:"posture_checks"`
	NetworkRoutes []NetworkRoute `yaml:"network_routes"`
	Users         []User         `yaml:"users"`
}
