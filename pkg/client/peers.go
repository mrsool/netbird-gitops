package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Instabug/netbird-gitops/pkg/data"
)

// ListPeers lists all NetBird peers
func (c Client) ListPeers(ctx context.Context) ([]data.Peer, error) {
	respBytes, err := c.doRequest(ctx, "GET", "peers", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPeers: %w", err)
	}
	var ret []data.Peer

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPeers: %w", err)
	}

	return ret, nil
}

// UpdatePeer updates a single NetBird user
func (c Client) UpdatePeer(ctx context.Context, peer data.Peer) error {
	if c.DryRun {
		return nil
	}

	body := map[string]interface{}{
		"name":                     peer.Name,
		"ssh_enabled":              peer.SSHEnabled,
		"login_expiration_enabled": !peer.ExpirationDisabled,
	}

	_, err := c.doRequest(ctx, "PUT", "peers/"+peer.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdatePeer: %w", err)
	}

	return nil
}
