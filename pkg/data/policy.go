package data

import (
	"errors"

	"github.com/Instabug/netbird-gitops/pkg/util"
)

// TODO: Make object conform to weird NetBird API

// Policy holds NetBird ACL Policy object
type Policy struct {
	ID                  string       `json:"id"`
	Name                string       `yaml:"name" json:"name"`
	Enabled             bool         `yaml:"enabled" json:"enabled"`
	Description         string       `yaml:"description" json:"description"`
	SourcePostureChecks []string     `yaml:"source_posture_checks"`
	Action              string       `yaml:"action"`
	Bidirectional       bool         `yaml:"bidirectional"`
	Protocol            string       `yaml:"protocol"`
	Sources             []string     `yaml:"sources"`
	Destinations        []string     `yaml:"destinations"`
	Rules               []PolicyRule `json:"rules"`
	Ports               []string     `yaml:"ports"`
}

// PolicyRule Policy.Rules section
type PolicyRule struct {
	SourceGroups      []Group  `json:"sources"`
	DestinationGroups []Group  `json:"destinations"`
	Description       string   `json:"description"`
	Action            string   `json:"action"`
	Bidirectional     bool     `json:"bidirectional"`
	Protocol          string   `json:"protocol"`
	Ports             []string `json:"ports"`
}

// Equals == operator
func (p Policy) Equals(o Policy) bool {
	return p.Name == o.Name &&
		p.Description == o.Description &&
		p.Enabled == o.Enabled &&
		util.SortedEqual(p.SourcePostureChecks, o.SourcePostureChecks) &&
		p.Action == o.Action &&
		p.Bidirectional == o.Bidirectional &&
		p.Protocol == o.Protocol &&
		util.SortedEqual(p.Sources, o.Sources) &&
		util.SortedEqual(p.Destinations, o.Destinations)
}

// Flatten converts uselessly nested policy rule to policy object
func (p *Policy) Flatten() error {
	if len(p.Rules) != 1 {
		return errors.New("Policy should have 1 rule exactly")
	}

	if p.Description == "" {
		p.Description = p.Rules[0].Description
	}
	p.Action = p.Rules[0].Action
	p.Bidirectional = p.Rules[0].Bidirectional
	p.Protocol = p.Rules[0].Protocol
	p.Ports = p.Rules[0].Ports
	p.Sources = util.Map(p.Rules[0].SourceGroups, func(g Group) string { return g.ID })
	p.Destinations = util.Map(p.Rules[0].DestinationGroups, func(g Group) string { return g.ID })
	return nil
}
