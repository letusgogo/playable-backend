package anbox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GatewayClient struct {
	config AnboxConfig
	client *http.Client
}

func NewGatewayClient(config AnboxConfig) *GatewayClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &GatewayClient{
		config: config,
		client: &http.Client{Transport: tr},
	}
}

// GetGatewayURL returns the gateway URL of the Anbox client
func (c *GatewayClient) GetGatewayURL() string {
	return c.config.Address
}

// GetAuthToken returns the authentication token
func (c *GatewayClient) GetAuthToken() string {
	return c.config.Token
}

// Create creates a new Anbox streaming session
func (c *GatewayClient) Create(ctx context.Context, req CreateSessionRequest) (*SessionDetails, error) {
	url := fmt.Sprintf("%s/1.0/sessions?api_token=%s", c.config.Address, c.config.Token)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", response.StatusCode, string(bodyBytes))
	}

	var result CreateSessionResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Metadata, nil
}

// CreateAsync creates a new Anbox streaming session asynchronously
func (c *GatewayClient) CreateAsync(ctx context.Context, req CreateSessionRequest) error {
	url := fmt.Sprintf("%s/1.0/sessions?api_token=%s", c.config.Address, c.config.Token)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(response.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", response.StatusCode, string(bodyBytes))
	}

	// We don't return the session details since it's async
	return nil
}

// Delete deletes an existing session
func (c *GatewayClient) Delete(ctx context.Context, sessionID string) error {
	url := fmt.Sprintf("%s/1.0/sessions/%s?api_token=%s", c.config.Address, sessionID, c.config.Token)

	request, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(response.Body)
		return fmt.Errorf("failed to delete session (status code: %d): %s", response.StatusCode, string(bodyBytes))
	}
	return nil
}
