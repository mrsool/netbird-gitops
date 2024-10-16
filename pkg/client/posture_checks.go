package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Instabug/netbird-gitops/pkg/data"
)

// ListPostureChecks lists all NetBird posture-checks
func (c Client) ListPostureChecks(ctx context.Context) ([]data.PostureCheck, error) {
	respBytes, err := c.doRequest(ctx, "GET", "posture-checks", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPostureChecks: %w", err)
	}
	var ret []data.PostureCheck

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListPostureChecks: %w", err)
	}

	return ret, nil
}

// UpdatePostureCheck updates a single NetBird postureCheck
func (c Client) UpdatePostureCheck(ctx context.Context, postureCheck data.PostureCheck) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	body, err := json.Marshal(postureCheck)
	if err != nil {
		return err
	}

	_, err = c.doRequest(ctx, "PUT", "posture-checks/"+postureCheck.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdatePostureCheck: %w", err)
	}
	return nil
}

// CreatePostureCheck updates a single NetBird postureCheck
func (c Client) CreatePostureCheck(ctx context.Context, postureCheck data.PostureCheck) (data.PostureCheck, error) {
	if c.DryRun {
		postureCheck.ID = postureCheck.Name
		return postureCheck, nil
	}

	body, err := json.Marshal(postureCheck)
	if err != nil {
		return data.PostureCheck{}, err
	}

	respBytes, err := c.doRequest(ctx, "POST", "posture-checks", body)
	if err != nil {
		return data.PostureCheck{}, fmt.Errorf("NetBird API: CreatePostureCheck: %w", err)
	}

	var ret data.PostureCheck

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return ret, fmt.Errorf("NetBird API: CreatePostureCheck: %w", err)
	}

	return ret, nil
}

// DeletePostureCheck updates a single NetBird postureCheck
func (c Client) DeletePostureCheck(ctx context.Context, postureCheck data.PostureCheck) error {
	if c.DryRun {
		slog.Info("DryRun==True")
		return nil
	}

	_, err := c.doRequest(ctx, "DELETE", "posture-checks/"+postureCheck.ID, nil)
	if err != nil {
		return fmt.Errorf("NetBird API: DeletePostureCheck: %w", err)
	}
	return nil
}
