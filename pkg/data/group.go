package data

// Group mapping of group ID and name
type Group struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Peers    []string `json:"-"`
	PeerData []Peer   `json:"peers"`
}
