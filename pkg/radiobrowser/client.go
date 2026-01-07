package radiobrowser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"
)

const (
	dnsDiscoveryHost = "all.api.radio-browser.info"
	userAgent        = "rig.fm/0.1.0"
	defaultTimeout   = 10 * time.Second
)

// Client is the Radio Browser API client
type Client struct {
	httpClient *http.Client
	servers    []string
	userAgent  string
}

// NewClient creates a new Radio Browser API client with server discovery
func NewClient() (*Client, error) {
	servers, err := discoverServers()
	if err != nil {
		return nil, fmt.Errorf("failed to discover servers: %w", err)
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers discovered")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		servers:   servers,
		userAgent: userAgent,
	}, nil
}

// NewClientWithServers creates a client with manually specified servers
func NewClientWithServers(servers []string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		servers:   servers,
		userAgent: userAgent,
	}
}

// discoverServers performs DNS lookup to find all available Radio Browser servers
func discoverServers() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", dnsDiscoveryHost)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed: %w", err)
	}

	servers := make([]string, 0, len(ips))
	for _, ip := range ips {
		// Reverse DNS to get the actual hostname
		names, err := net.LookupAddr(ip.String())
		if err == nil && len(names) > 0 {
			// Remove trailing dot from hostname
			hostname := names[0]
			if len(hostname) > 0 && hostname[len(hostname)-1] == '.' {
				hostname = hostname[:len(hostname)-1]
			}
			servers = append(servers, fmt.Sprintf("https://%s", hostname))
		} else {
			// Fallback to IP if reverse DNS fails
			servers = append(servers, fmt.Sprintf("https://%s", ip.String()))
		}
	}

	// Randomize server order for load balancing
	rand.Shuffle(len(servers), func(i, j int) {
		servers[i], servers[j] = servers[j], servers[i]
	})

	return servers, nil
}

// get performs a GET request to the API with automatic server fallback
func (c *Client) get(endpoint string) ([]byte, error) {
	var lastErr error

	for _, server := range c.servers {
		url := fmt.Sprintf("%s%s", server, endpoint)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		return body, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all servers failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("all servers failed with no error")
}

// unmarshalJSON is a helper to unmarshal JSON responses
func unmarshalJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// GetServers returns the list of discovered servers
func (c *Client) GetServers() []string {
	return c.servers
}
