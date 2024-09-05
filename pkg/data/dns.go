package data

import (
	"slices"

	"github.com/Instabug/netbird-gitops/pkg/util"
)

// DNS holds NetBird DNS Management settings
type DNS struct {
	DisableFor []string `yaml:"disableFor" json:"disabled_management_groups"`
}

// DNSResponse holds NetBird DNS Management Settings Response
type DNSResponse struct {
	Items struct {
		DisableFor []string `json:"disabled_management_groups"`
	} `json:"items"`
}

// Nameserver holds one nameserver group settings
type Nameserver struct {
	ID                   string             `json:"id"`
	Name                 string             `yaml:"name" json:"name"`
	Description          string             `yaml:"description" json:"description"`
	Nameservers          []NameserverServer `yaml:"nameservers" json:"nameservers"`
	Enabled              bool               `yaml:"enabled" json:"enabled"`
	Groups               []string           `yaml:"groups" json:"groups"`
	Primary              bool               `yaml:"primary" json:"primary"`
	Domains              []string           `yaml:"domains" json:"domains"`
	SearchDomainsEnabled bool               `yaml:"search_domains_enabled" json:"search_domains_enabled"`
}

// NameserverServer holds prot://ip:port
type NameserverServer struct {
	IP     string `yaml:"ip"`
	NSType string `yaml:"ns_type"`
	Port   uint   `yaml:"port"`
}

// Equals == operator
func (ns Nameserver) Equals(o Nameserver) bool {
	return ns.ID == o.ID &&
		ns.Name == o.Name &&
		ns.Description == o.Description &&
		slices.EqualFunc(ns.Nameservers, o.Nameservers, func(a, b NameserverServer) bool {
			return a.IP == b.IP && a.NSType == b.NSType && a.Port == b.Port
		}) &&
		ns.Enabled == o.Enabled &&
		util.SortedEqual(ns.Groups, o.Groups) &&
		ns.Primary == o.Primary &&
		util.SortedEqual(ns.Domains, o.Domains) &&
		ns.SearchDomainsEnabled == o.SearchDomainsEnabled
}
