package anbox

import (
	"context"
)

// Client implements the AnboxClient interface by combining GatewayClient and AMSClient
type Client struct {
	gatewayClient *GatewayClient
	amsClient     *AMSClient
}

// NewClient creates a new Anbox client with both Gateway and AMS clients
func NewClient(cfg AnboxConfig) (*Client, error) {
	amsClient, err := NewAMSClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		gatewayClient: NewGatewayClient(cfg),
		amsClient:     amsClient,
	}, nil
}

// CreateAsync creates a new Anbox streaming session asynchronously
func (c *Client) CreateAsync(ctx context.Context, req CreateSessionRequest) error {
	return c.gatewayClient.CreateAsync(ctx, req)
}

// Delete deletes an existing session
func (c *Client) Delete(ctx context.Context, sessionID string) error {
	return c.gatewayClient.Delete(ctx, sessionID)
}

// GetAllRunningSession gets all running sessions from AMS
func (c *Client) GetAllRunningSession(ctx context.Context) ([]*SessionDetails, error) {
	return c.amsClient.GetAllRunningSession(ctx)
}

// GetGatewayURL returns the gateway URL
func (c *Client) GetGatewayURL() string {
	return c.gatewayClient.GetGatewayURL()
}

// GetAuthToken returns the authentication token
func (c *Client) GetAuthToken() string {
	return c.gatewayClient.GetAuthToken()
}
