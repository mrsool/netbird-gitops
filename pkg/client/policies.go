package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Instabug/netbird-gitops/pkg/data"
)

// ListPolicies lists all NetBird policies
func (c Client) ListPolicies(ctx context.Context) ([]data.Policy, error) {
	respBytes, err := c.doRequest(ctx, "GET", "policies", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPolicies: %w", err)
	}
	var ret []data.Policy

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPolicies: %w", err)
	}

	for idx := range ret {
		ret[idx].Flatten()
	}

	return ret, nil
}

// UpdatePolicy updates a single NetBird policy
func (c Client) UpdatePolicy(ctx context.Context, policy data.Policy) error {
	if c.DryRun {
		return nil
	}

	body := map[string]interface{}{
		"name":                  policy.Name,
		"description":           policy.Description,
		"enabled":               policy.Enabled,
		"source_posture_checks": policy.SourcePostureChecks,
		"rules": []map[string]interface{}{
			{
				"name":          policy.Name,
				"description":   policy.Description,
				"enabled":       policy.Enabled,
				"action":        policy.Action,
				"bidirectional": policy.Bidirectional,
				"protocol":      policy.Protocol,
				"ports":         policy.Ports,
				"sources":       policy.Sources,
				"destinations":  policy.Destinations,
			},
		},
	}

	_, err := c.doRequest(ctx, "PUT", "policies/"+policy.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdatePolicy: %w", err)
	}
	return nil
}

// CreatePolicy updates a single NetBird policy
func (c Client) CreatePolicy(ctx context.Context, policy data.Policy) error {
	if c.DryRun {
		return nil
	}

	body := map[string]interface{}{
		"name":                  policy.Name,
		"description":           policy.Description,
		"enabled":               policy.Enabled,
		"source_posture_checks": policy.SourcePostureChecks,
		"rules": []map[string]interface{}{
			{
				"name":          policy.Name,
				"description":   policy.Description,
				"enabled":       policy.Enabled,
				"action":        policy.Action,
				"bidirectional": policy.Bidirectional,
				"protocol":      policy.Protocol,
				"ports":         policy.Ports,
				"sources":       policy.Sources,
				"destinations":  policy.Destinations,
			},
		},
	}

	_, err := c.doRequest(ctx, "POST", "policies", body)
	if err != nil {
		return fmt.Errorf("NetBird API: CreatePolicy: %w", err)
	}
	return nil
}

// DeletePolicy updates a single NetBird policy
func (c Client) DeletePolicy(ctx context.Context, policy data.Policy) error {
	if c.DryRun {
		return nil
	}

	_, err := c.doRequest(ctx, "DELETE", "policies/"+policy.ID, nil)
	if err != nil {
		return fmt.Errorf("NetBird API: DeletePolicy: %w", err)
	}
	return nil
}
