package anbox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealGatewayClient(t *testing.T) {
	client := NewGatewayClient(AnboxConfig{
		Address: "https://dev.android.gateway.gamingnow.co:4000",
		Token:   "AgEUYW5ib3gtc3RyZWFtLWdhdGV3YXkCCmRldi1jbGllbnQAAhQyMDI1LTA3LTI0VDAyOjIwOjM4WgACFDIwMjYtMDctMjRUMDI6MjA6MzhaAAAGIPLA63vBcqpWlVfGPkC6_GFIipnLtN7HHVTEZ1nadfvb",
	})
	ctx := context.Background()
	req := CreateSessionRequest{
		App:         "idle_weapon",
		AppVersion:  1,
		Ephemeral:   true,
		ExtraData:   "test-data",
		IdleTimeMin: 5,
		Joinable:    true,
		Screen: Screen{
			Width:   720,
			Height:  1240,
			Density: 320,
			FPS:     30,
		},
	}
	session, err := client.Create(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	t.Logf("Session created: %+v", session)
}

func TestNewGatewayClient(t *testing.T) {
	client := NewGatewayClient(AnboxConfig{
		Address: "https://gateway.example.com",
		Token:   "test-token",
	})

	if client.GetGatewayURL() != "https://gateway.example.com" {
		t.Errorf("Expected gateway URL 'https://gateway.example.com', got '%s'", client.GetGatewayURL())
	}

	if client.GetAuthToken() != "test-token" {
		t.Errorf("Expected auth token 'test-token', got '%s'", client.GetAuthToken())
	}

	if client.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestGetGatewayURL(t *testing.T) {
	client := NewGatewayClient(AnboxConfig{
		Address: "https://gateway.example.com",
		Token:   "test-token",
	})

	url := client.GetGatewayURL()
	if url != "https://gateway.example.com" {
		t.Errorf("Expected gateway URL 'https://gateway.example.com', got '%s'", url)
	}
}

func TestCreateSession_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a POST request to the correct endpoint
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/1.0/sessions" {
			t.Errorf("Expected path '/1.0/sessions', got '%s'", r.URL.Path)
		}

		// Check if API token is present
		if r.URL.Query().Get("api_token") != "test-token" {
			t.Error("Expected api_token in query parameters")
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"type": "sync",
			"status": "Success",
			"status_code": 200,
			"metadata": {
				"id": "test-session-id",
				"region": "us-east-1",
				"url": "wss://gateway.example.com/sessions/test-session-id",
				"stun_servers": [
					{
						"urls": ["stun:stun.example.com:3478"]
					}
				],
				"status": "running",
				"joinable": true
			}
		}`))
	}))
	defer server.Close()

	// Debug: 打印实际的服务器URL
	t.Logf("Mock server running at: %s", server.URL)

	// Create client with test server URL
	client := NewGatewayClient(AnboxConfig{
		Address: server.URL,
		Token:   "test-token",
	})

	// Create session request
	req := CreateSessionRequest{
		App:         "test-app",
		AppVersion:  1,
		Ephemeral:   true,
		ExtraData:   "test-data",
		IdleTimeMin: 30,
		Joinable:    true,
		Screen: Screen{
			Width:   1920,
			Height:  1080,
			Density: 240,
			FPS:     60,
		},
	}

	// Create session
	ctx := context.Background()
	session, err := client.Create(ctx, req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if session.ID != "test-session-id" {
		t.Errorf("Expected session ID 'test-session-id', got '%s'", session.ID)
	}

	if session.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", session.Region)
	}

	if session.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", session.Status)
	}

	if !session.Joinable {
		t.Error("Expected session to be joinable")
	}
}

func TestDeleteSession_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a DELETE request to the correct endpoint
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		if r.URL.Path != "/1.0/sessions/test-session-id" {
			t.Errorf("Expected path '/1.0/sessions/test-session-id', got '%s'", r.URL.Path)
		}

		// Check if API token is present
		if r.URL.Query().Get("api_token") != "test-token" {
			t.Error("Expected api_token in query parameters")
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewGatewayClient(AnboxConfig{
		Address: server.URL,
		Token:   "test-token",
	})

	// Delete session
	ctx := context.Background()
	err := client.Delete(ctx, "test-session-id")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
