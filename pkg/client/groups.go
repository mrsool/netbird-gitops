package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Instabug/netbird-gitops/pkg/data"
)

// ListGroups lists all NetBird groups
func (c Client) ListGroups(ctx context.Context) ([]data.Group, error) {
	respBytes, err := c.doRequest(ctx, "GET", "groups", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListGroups: %w", err)
	}
	var ret []data.Group

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListGroups: %w", err)
	}

	return ret, nil
}

// CreateGroup create NetBird Group
func (c Client) CreateGroup(ctx context.Context, group data.Group) (data.Group, error) {
	if c.DryRun {
		group.ID = group.Name
		return group, nil
	}
	body := map[string]interface{}{
		"name": group.Name,
	}

	respBytes, err := c.doRequest(ctx, "POST", "groups", body)
	if err != nil {
		return data.Group{}, fmt.Errorf("NetBird API: CreateGroup: %w", err)
	}

	var ret data.Group

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return ret, fmt.Errorf("NetBird API: CreateGroup: %w", err)
	}

	return ret, nil
}

// UpdateGroup update NetBird Group
func (c Client) UpdateGroup(ctx context.Context, group data.Group) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}
	body := map[string]interface{}{
		"name":  group.Name,
		"peers": group.Peers,
	}

	_, err := c.doRequest(ctx, "PUT", "groups/"+group.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: CreateGroup: %w", err)
	}

	return nil
}

// DeleteGroup delete NetBird Group
func (c Client) DeleteGroup(ctx context.Context, group data.Group) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	_, err := c.doRequest(ctx, "DELETE", "groups/"+group.ID, nil)
	if err != nil {
		return fmt.Errorf("NetBird API: DeleteGroup: %w", err)
	}

	return nil
}
