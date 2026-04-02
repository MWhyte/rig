package radiobrowser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiHost        = "all.api.radio-browser.info"
	userAgent      = "rig.fm/0.1.0"
	defaultTimeout = 10 * time.Second
)

// Client is the Radio Browser API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewClient creates a new Radio Browser API client.
func NewClient() (*Client, error) {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    fmt.Sprintf("https://%s", apiHost),
		userAgent:  userAgent,
	}, nil
}

// get performs a GET request to the API.
func (c *Client) get(endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// unmarshalJSON is a helper to unmarshal JSON responses.
func unmarshalJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}
