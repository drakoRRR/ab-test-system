package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type httpClient interface {
	FetchConfig(ctx context.Context, apiKey string) (*SDKConfig, error)
	SendEvents(ctx context.Context, apiKey string, batch []event) error
}

type defaultHTTPClient struct {
	baseURL string
	client  *http.Client
}

func newDefaultHTTPClient(baseURL string, client *http.Client) *defaultHTTPClient {
	return &defaultHTTPClient{baseURL: baseURL, client: client}
}

func (c *defaultHTTPClient) FetchConfig(ctx context.Context, apiKey string) (*SDKConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/sdk/config", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var cfg SDKConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return &cfg, nil
}

func (c *defaultHTTPClient) SendEvents(ctx context.Context, apiKey string, batch []event) error {
	payload := struct {
		Events []event `json:"events"`
	}{Events: batch}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sdk/events", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, b)
	}
	return nil
}
