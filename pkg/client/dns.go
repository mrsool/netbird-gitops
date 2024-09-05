package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Instabug/netbird-gitops/pkg/data"
)

// GetDNSSettings Get NetBird DNS settings
func (c Client) GetDNSSettings(ctx context.Context) (data.DNSResponse, error) {
	respBytes, err := c.doRequest(ctx, "GET", "dns/settings", nil)
	if err != nil {
		return data.DNSResponse{}, fmt.Errorf("NetBird API: GetDNSSettings: %w", err)
	}
	var ret data.DNSResponse

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return data.DNSResponse{}, fmt.Errorf("NetBird API: GetDNSSettings: %w", err)
	}

	return ret, nil
}

// UpdateDNSSettings Update NetBird DNS settings
func (c Client) UpdateDNSSettings(ctx context.Context, settings data.DNS) error {
	if c.DryRun {
		return nil
	}

	body, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.doRequest(ctx, "POST", "dns/settings", body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdateDNSSettings: %w", err)
	}

	return nil
}

// ListNameservers List DNS Nameservers
func (c Client) ListNameservers(ctx context.Context) ([]data.Nameserver, error) {
	respBytes, err := c.doRequest(ctx, "GET", "dns/nameservers", nil)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListNameservers: %w", err)
	}
	var ret []data.Nameserver

	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("NetBird API: ListNameservers: %w", err)
	}

	return ret, nil
}

// UpdateNameserver updates a single NetBird nameserver
func (c Client) UpdateNameserver(ctx context.Context, nameserver data.Nameserver) error {
	if c.DryRun {
		return nil
	}

	body, err := json.Marshal(nameserver)
	if err != nil {
		return err
	}

	_, err = c.doRequest(ctx, "PUT", "dns/nameservers/"+nameserver.ID, body)
	if err != nil {
		return fmt.Errorf("NetBird API: UpdateNameserver: %w", err)
	}
	return nil
}

// CreateNameserver updates a single NetBird nameserver
func (c Client) CreateNameserver(ctx context.Context, nameserver data.Nameserver) error {
	if c.DryRun {
		nameserver.ID = nameserver.Name
		return nil
	}

	body, err := json.Marshal(nameserver)
	if err != nil {
		return err
	}

	_, err = c.doRequest(ctx, "POST", "dns/nameservers", body)
	if err != nil {
		return fmt.Errorf("NetBird API: CreateNameserver: %w", err)
	}

	return nil
}

// DeleteNameserver updates a single NetBird nameserver
func (c Client) DeleteNameserver(ctx context.Context, nameserver data.Nameserver) error {
	if c.DryRun {
		return nil
	}

	_, err := c.doRequest(ctx, "DELETE", "dns/nameservers/"+nameserver.ID, nil)
	if err != nil {
		return fmt.Errorf("NetBird API: DeleteNameserver: %w", err)
	}
	return nil
}
