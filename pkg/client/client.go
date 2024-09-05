package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Client NetBird API Client
type Client struct {
	managementAPI string
	token         string
	client        *http.Client
	DryRun        bool
}

// NewClient returns a new NetBird API Client
func NewClient(managementAPI, token string, dryRun bool) *Client {
	managementAPI = strings.TrimSuffix(managementAPI, "/")
	return &Client{
		managementAPI: managementAPI,
		token:         token,
		client:        http.DefaultClient,
		DryRun:        dryRun,
	}
}

func (c Client) doRequest(ctx context.Context, method, resource string, body interface{}) ([]byte, error) {
	t1 := time.Now()
	slog.Info(method+" /api/"+resource, "body", body)
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.managementAPI+"/api/"+resource, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Token "+c.token)
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	slog.Debug(method+" /api/"+resource, "response", string(respBytes), "time", time.Since(t1))
	slog.Info(method+" /api/"+resource, "response_code", resp.StatusCode, "time", time.Since(t1), "content_size", len(respBytes))

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	return respBytes, nil
}
