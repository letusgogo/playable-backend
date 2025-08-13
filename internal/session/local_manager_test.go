package session

import (
	"context"
	"testing"
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
)

// MockAnboxClient for testing
type MockAnboxClient struct {
	sessions    map[string]bool
	createError error
	deleteError error
}

func NewMockAnboxClient() *MockAnboxClient {
	return &MockAnboxClient{
		sessions: make(map[string]bool),
	}
}

func (m *MockAnboxClient) CreateAsync(ctx context.Context, req anbox.CreateSessionRequest) error {
	return m.createError
}

func (m *MockAnboxClient) Delete(ctx context.Context, sessionID string) error {
	delete(m.sessions, sessionID)
	return m.deleteError
}

func (m *MockAnboxClient) GetAllRunningSession(ctx context.Context) ([]*anbox.SessionDetails, error) {
	var sessions []*anbox.SessionDetails
	for id := range m.sessions {
		sessions = append(sessions, &anbox.SessionDetails{
			ID:     id,
			Status: "running",
		})
	}
	return sessions, nil
}

func (m *MockAnboxClient) GetGatewayURL() string {
	return "mock://gateway"
}

func (m *MockAnboxClient) GetAuthToken() string {
	return "mock-token"
}

func TestLocalSessionManager_StateTransitions(t *testing.T) {
	cfg := &Config{
		GameName:         "test-game",
		Min:              1,
		Max:              10,
		SessionTTL:       5 * time.Minute,
		HeartbeatTimeout: 1 * time.Minute,
		SyncInterval:     10 * time.Second,
		ScreenConfig: &ScreenConfig{
			Width:   720,
			Height:  1240,
			Density: 320,
			Fps:     30,
		},
	}

	mockClient := NewMockAnboxClient()
	manager := NewLocalSessionManager(cfg, mockClient)

	ctx := context.Background()

	// Test: Create a session manually and add to cache
	session := &Session{
		ID:            "test-session-1",
		Game:          cfg.GameName,
		Status:        Cold,
		GatewayURL:    mockClient.GetGatewayURL(),
		AuthToken:     mockClient.GetAuthToken(),
		ExpiresAt:     time.Now().Add(cfg.SessionTTL),
		LastHeartbeat: time.Now(),
		CreatedAt:     time.Now(),
	}

	manager.mu.Lock()
	manager.cache[session.ID] = session
	manager.mu.Unlock()

	// Test: AcquireCold (cold -> warming)
	coldSession, err := manager.AcquireCold(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire cold session: %v", err)
	}

	if coldSession.Status != Warming {
		t.Errorf("Expected session status to be Warming, got %s", coldSession.Status)
	}

	// Test: SetWarmed (warming -> warmed)
	err = manager.SetWarmed(ctx, coldSession.ID)
	if err != nil {
		t.Fatalf("Failed to set session as warmed: %v", err)
	}

	// Verify session status changed
	retrievedSession, err := manager.GetSession(ctx, coldSession.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.Status != Warmed {
		t.Errorf("Expected session status to be Warmed, got %s", retrievedSession.Status)
	}

	// Test: AcquireWarmed (warmed -> in_use)
	warmedSession, err := manager.AcquireWarmed(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire warmed session: %v", err)
	}

	if warmedSession.ID != coldSession.ID {
		t.Errorf("Expected same session ID, got different sessions")
	}

	if warmedSession.Status != InUse {
		t.Errorf("Expected session status to be InUse, got %s", warmedSession.Status)
	}

	// Test: Heartbeat
	err = manager.Heartbeat(ctx, warmedSession.ID)
	if err != nil {
		t.Fatalf("Failed to send heartbeat: %v", err)
	}

	// Test: Release (delete session)
	err = manager.Release(ctx, warmedSession.ID)
	if err != nil {
		t.Fatalf("Failed to release session: %v", err)
	}

	// Verify session was deleted
	_, err = manager.GetSession(ctx, warmedSession.ID)
	if err == nil {
		t.Errorf("Expected session to be deleted, but it still exists")
	}
}

func TestLocalSessionManager_PoolStatus(t *testing.T) {
	cfg := &Config{
		GameName:         "test-game",
		Min:              1,
		Max:              10,
		SessionTTL:       5 * time.Minute,
		HeartbeatTimeout: 1 * time.Minute,
		SyncInterval:     10 * time.Second,
		ScreenConfig: &ScreenConfig{
			Width:   720,
			Height:  1240,
			Density: 320,
			Fps:     30,
		},
	}

	mockClient := NewMockAnboxClient()
	manager := NewLocalSessionManager(cfg, mockClient)

	ctx := context.Background()

	// Add sessions with different statuses
	sessions := []*Session{
		{ID: "cold-1", Status: Cold, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "cold-2", Status: Cold, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "warming-1", Status: Warming, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "warmed-1", Status: Warmed, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "inuse-1", Status: InUse, LastHeartbeat: time.Now(), CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)},
	}

	manager.mu.Lock()
	for _, session := range sessions {
		manager.cache[session.ID] = session
	}
	manager.mu.Unlock()

	// Test: PoolStatus
	status, err := manager.PoolStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get pool status: %v", err)
	}

	if status.Total != 5 {
		t.Errorf("Expected total 5 sessions, got %d", status.Total)
	}

	if status.Cold != 2 {
		t.Errorf("Expected 2 cold sessions, got %d", status.Cold)
	}

	if status.Warming != 1 {
		t.Errorf("Expected 1 warming session, got %d", status.Warming)
	}

	if status.Warmed != 1 {
		t.Errorf("Expected 1 warmed session, got %d", status.Warmed)
	}

	if status.InUse != 1 {
		t.Errorf("Expected 1 in-use session, got %d", status.InUse)
	}
}

func TestLocalSessionManager_ListSessions(t *testing.T) {
	cfg := &Config{
		GameName:         "test-game",
		Min:              1,
		Max:              10,
		SessionTTL:       5 * time.Minute,
		HeartbeatTimeout: 1 * time.Minute,
		SyncInterval:     10 * time.Second,
		ScreenConfig: &ScreenConfig{
			Width:   720,
			Height:  1240,
			Density: 320,
			Fps:     30,
		},
	}

	mockClient := NewMockAnboxClient()
	manager := NewLocalSessionManager(cfg, mockClient)

	ctx := context.Background()

	// Add sessions with different statuses
	sessions := []*Session{
		{ID: "cold-1", Status: Cold, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "cold-2", Status: Cold, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
		{ID: "warmed-1", Status: Warmed, LastHeartbeat: time.Now(), CreatedAt: time.Now()},
	}

	manager.mu.Lock()
	for _, session := range sessions {
		manager.cache[session.ID] = session
	}
	manager.mu.Unlock()

	// Test: ListSessions for Cold status
	coldSessions, err := manager.ListSessions(ctx)
	if err != nil {
		t.Fatalf("Failed to list cold sessions: %v", err)
	}

	if len(coldSessions) != 2 {
		t.Errorf("Expected 2 cold sessions, got %d", len(coldSessions))
	}

	// Test: ListSessions for Warmed status
	warmedSessions, err := manager.ListSessions(ctx)
	if err != nil {
		t.Fatalf("Failed to list warmed sessions: %v", err)
	}

	if len(warmedSessions) != 1 {
		t.Errorf("Expected 1 warmed session, got %d", len(warmedSessions))
	}
}

func TestLocalSessionManager_ErrorHandling(t *testing.T) {
	cfg := &Config{
		GameName:         "test-game",
		Min:              1,
		Max:              10,
		SessionTTL:       5 * time.Minute,
		HeartbeatTimeout: 1 * time.Minute,
		SyncInterval:     10 * time.Second,
		ScreenConfig: &ScreenConfig{
			Width:   720,
			Height:  1240,
			Density: 320,
			Fps:     30,
		},
	}

	mockClient := NewMockAnboxClient()
	manager := NewLocalSessionManager(cfg, mockClient)

	ctx := context.Background()

	// Test: AcquireCold when no cold sessions available
	_, err := manager.AcquireCold(ctx)
	if err == nil {
		t.Errorf("Expected error when no cold sessions available, but got none")
	}

	// Test: SetWarmed with non-existent session ID
	err = manager.SetWarmed(ctx, "non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent session, but got none")
	}

	// Test: AcquireWarmed when no warmed sessions available
	_, err = manager.AcquireWarmed(ctx)
	if err == nil {
		t.Errorf("Expected error when no warmed sessions available, but got none")
	}

	// Test: GetSession with non-existent ID
	_, err = manager.GetSession(ctx, "non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent session, but got none")
	}
}
