package wiremock

// todo effectively this whole class can be replaced with the client at https://github.com/wiremock/go-wiremock
// see also https://wiremock.org/docs/solutions/golang/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// StubMapping represents a Wiremock stub configuration
type StubMapping struct {
	Request  RequestPattern `json:"request"`
	Response ResponseDef    `json:"response"`
}

// RequestPattern defines the request matching criteria
type RequestPattern struct {
	Method string `json:"method"`
	URL    string `json:"url,omitempty"`
	Path   string `json:"urlPath,omitempty"`
}

// ResponseDef defines the response to return
type ResponseDef struct {
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Status  int               `json:"status"`
}

// Client handles interactions with the Wiremock admin API
type Client struct {
	HTTPClient *http.Client
	AdminURL   string
}

// NewClient creates a new Wiremock admin client
func NewClient(adminURL string) *Client {
	return &Client{
		AdminURL:   adminURL,
		HTTPClient: &http.Client{},
	}
}

// CreateStub creates a new stub mapping in Wiremock
func (c *Client) CreateStub(stub StubMapping) error {
	data, err := json.Marshal(stub)
	if err != nil {
		return fmt.Errorf("failed to marshal stub: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.AdminURL+"/__admin/mappings",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create stub: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusCreated {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("unexpected status %d (failed to read response body: %w)", resp.StatusCode, readErr)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Reset removes all stub mappings from Wiremock
func (c *Client) Reset() error {
	req, err := http.NewRequest(
		http.MethodPost,
		c.AdminURL+"/__admin/reset",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create reset request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reset wiremock: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("unexpected status %d (failed to read response body: %w)", resp.StatusCode, readErr)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// HealthCheck verifies Wiremock is available
func (c *Client) HealthCheck() error {
	req, err := http.NewRequest(
		http.MethodGet,
		c.AdminURL+"/__admin/health",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("wiremock health check failed: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wiremock unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

func closeBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Printf("wiremock client: failed to close response body: %v", err)
	}
}
