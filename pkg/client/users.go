package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mrsool/netbird-gitops/pkg/data"
)

// ListUsers lists all NetBird users
func (c Client) ListUsers(ctx context.Context) ([]data.User, error) {
	respBytes, err := c.doRequest(ctx, "GET", "users", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListUsers: %w", err)
	}
	var ret []data.User

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListUsers: %w", err)
	}

	return ret, nil
}

// UpdateUser updates a single NetBird user
func (c Client) UpdateUser(ctx context.Context, user data.User) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	body := map[string]interface{}{
		"role":        user.GetRole(),
		"auto_groups": user.Groups,
		"is_blocked":  user.Blocked,
	}

	_, err := c.doRequest(ctx, "PUT", "users/"+user.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdateUser: %w", err)
	}
	return nil
}
