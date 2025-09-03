package data

import "github.com/mrsool/netbird-gitops/pkg/util"

// NetworkRoute NetBird network route object
type NetworkRoute struct {
	ID          string   `json:"id"`
	NetworkType string   `yaml:"network_type" json:"network_type"`
	Description string   `yaml:"description" json:"description"`
	NetworkID   string   `yaml:"network_id" json:"network_id"`
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Peer        string   `yaml:"peer" json:"peer"`
	PeerGroups  []string `yaml:"peer_groups" json:"peer_groups"`
	Network     string   `yaml:"network" json:"network"`
	Domains     []string `yaml:"domains" json:"domains"`
	Metric      int      `yaml:"metric" json:"metric"`
	Masquerade  bool     `yaml:"masquerade" json:"masquerade"`
	Groups      []string `yaml:"groups" json:"groups"`
	KeepRoute   bool     `yaml:"keep_route" json:"keep_route"`
}

// Equals returns if network routes are equal
func (n NetworkRoute) Equals(o NetworkRoute) bool {
	return n.NetworkType == o.NetworkType &&
		n.Description == o.Description &&
		n.NetworkID == o.NetworkID &&
		n.Enabled == o.Enabled &&
		n.Peer == o.Peer &&
		util.SortedEqual(n.PeerGroups, o.PeerGroups) &&
		((len(n.Domains) != 0 || len(o.Domains) != 0) || n.Network == o.Network) &&
		util.SortedEqual(n.Domains, o.Domains) &&
		n.Metric == o.Metric &&
		n.Masquerade == o.Masquerade &&
		util.SortedEqual(n.Groups, o.Groups) &&
		n.KeepRoute == o.KeepRoute
}
