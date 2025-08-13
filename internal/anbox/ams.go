package anbox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AMSClient handles communication with Anbox Management Service
type AMSClient struct {
	cfg    *AnboxConfig
	client *http.Client
}

// NewAMSClient creates a new AMS client with certificate authentication
func NewAMSClient(config AnboxConfig) (*AMSClient, error) {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair(config.AmsCert, config.AmsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // Skip verification as we're using self-signed certs
	}

	// Create HTTP client with TLS config
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Ensure baseURL has https:// scheme and no trailing slash
	baseURL := config.AmsAddr
	if !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Update config with normalized URL
	config.AmsAddr = baseURL

	return &AMSClient{
		cfg:    &config,
		client: httpClient,
	}, nil
}

// GetAllRunningSession gets all running sessions from AMS
func (a *AMSClient) GetAllRunningSession(ctx context.Context) ([]*SessionDetails, error) {
	url := fmt.Sprintf("%s/1.0/instances", a.cfg.AmsAddr)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", response.StatusCode, string(bodyBytes))
	}

	var result ListInstancesResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract instance IDs from metadata paths
	var sessions []*SessionDetails
	for _, path := range result.Metadata {
		// Extract ID from path "/1.0/instances/instance-id"
		instanceID := strings.TrimPrefix(path, "/1.0/instances/")
		if instanceID != "" {
			// Get detailed information for each instance to check if it's running
			details, err := a.GetInstanceDetails(ctx, instanceID)
			if err != nil {
				// Continue with other instances if one fails
				continue
			}

			// Only include running instances
			if details.Status == "running" {
				// Try to extract session ID from tags or use instance ID
				sessionID := instanceID
				if extractedID := GetSessionIDFromTags(details.Tags); extractedID != "" {
					sessionID = extractedID
				}

				session := &SessionDetails{
					ID:     sessionID,
					Status: details.Status,
					// Map other fields as needed
					Region:   "", // AMS doesn't provide region info
					URL:      "", // This would come from gateway
					Joinable: true,
				}
				sessions = append(sessions, session)
			}
		}
	}

	return sessions, nil
}

// ListInstances retrieves all instances from AMS
func (a *AMSClient) ListInstances(ctx context.Context) (*ListInstanceDetails, error) {
	url := fmt.Sprintf("%s/1.0/instances", a.cfg.AmsAddr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rawResponse ListInstancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract instance IDs from metadata paths
	instanceIDs := make([]string, 0, len(rawResponse.Metadata))
	for _, path := range rawResponse.Metadata {
		// Extract ID from path "/1.0/instances/instance-id"
		id := strings.TrimPrefix(path, "/1.0/instances/")
		if id != "" {
			instanceIDs = append(instanceIDs, id)
		}
	}

	return &ListInstanceDetails{
		InstanceIDs: instanceIDs,
		TotalCount:  rawResponse.TotalSize,
	}, nil
}

// GetInstanceDetails retrieves detailed information about a specific instance
func (a *AMSClient) GetInstanceDetails(ctx context.Context, instanceID string) (*InstanceDetails, error) {
	// Extract the actual instance ID from the full path if necessary
	instanceID = strings.TrimPrefix(instanceID, "/1.0/instances/")

	url := fmt.Sprintf("%s/1.0/instances/%s", a.cfg.AmsAddr, instanceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result InstanceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Metadata, nil
}

// GetSessionIDFromTags extracts the session ID from instance tags
func GetSessionIDFromTags(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "session=") {
			// assuming there is only one session id in the tags
			return strings.TrimPrefix(tag, "session=")
		}
	}
	return ""
}
