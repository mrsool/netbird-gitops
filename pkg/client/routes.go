package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mrsool/netbird-gitops/pkg/data"
)

// ListNetworkRoutes lists all NetBird routes
func (c Client) ListNetworkRoutes(ctx context.Context) ([]data.NetworkRoute, error) {
	respBytes, err := c.doRequest(ctx, "GET", "routes", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListNetworkRoutes: %w", err)
	}
	var ret []data.NetworkRoute

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListNetworkRoutes: %w", err)
	}

	return ret, nil
}

// UpdateNetworkRoute updates a single NetBird route
func (c Client) UpdateNetworkRoute(ctx context.Context, route data.NetworkRoute) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	body := map[string]interface{}{
		"description": route.Description,
		"network_id":  route.NetworkID,
		"enabled":     route.Enabled,
		"metric":      route.Metric,
		"masquerade":  route.Masquerade,
		"groups":      route.Groups,
		"keep_route":  route.KeepRoute,
	}

	if len(route.Domains) > 0 {
		body["domains"] = route.Domains
	} else {
		body["network"] = route.Network
	}

	if len(route.PeerGroups) > 0 {
		body["peer_groups"] = route.PeerGroups
	} else {
		body["peer"] = route.Peer
	}

	_, err := c.doRequest(ctx, "PUT", "routes/"+route.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdateNetworkRoute: %w", err)
	}
	return nil
}

// CreateNetworkRoute updates a single NetBird route
func (c Client) CreateNetworkRoute(ctx context.Context, route data.NetworkRoute) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	body := map[string]interface{}{
		"description": route.Description,
		"network_id":  route.NetworkID,
		"enabled":     route.Enabled,
		"metric":      route.Metric,
		"masquerade":  route.Masquerade,
		"groups":      route.Groups,
		"keep_route":  route.KeepRoute,
	}

	if len(route.Domains) > 0 {
		body["domains"] = route.Domains
	} else {
		body["network"] = route.Network
	}

	if len(route.PeerGroups) > 0 {
		body["peer_groups"] = route.PeerGroups
	} else {
		body["peer"] = route.Peer
	}

	_, err := c.doRequest(ctx, "POST", "routes", body)
	if err != nil {
		return fmt.Errorf("NetBird API: CreateNetworkRoute: %w", err)
	}
	return nil
}

// DeleteNetworkRoute updates a single NetBird route
func (c Client) DeleteNetworkRoute(ctx context.Context, route data.NetworkRoute) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	_, err := c.doRequest(ctx, "DELETE", "routes/"+route.ID, nil)
	if err != nil {
		return fmt.Errorf("NetBird API: DeleteNetworkRoute: %w", err)
	}
	return nil
}
