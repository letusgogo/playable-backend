package session

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
	"github.com/letusgogo/quick/logger"
)

type LocalSessionManager struct {
	mu          sync.RWMutex
	cache       map[string]*Session
	anboxClient AnboxClient
	cfg         *Config
	syncStopCh  chan struct{}
	started     bool
}

func NewLocalSessionManager(cfg *Config, anboxClient AnboxClient) *LocalSessionManager {
	return &LocalSessionManager{
		cache:       make(map[string]*Session),
		anboxClient: anboxClient,
		cfg:         cfg,
		syncStopCh:  make(chan struct{}),
	}
}

// Init initializes the session manager with configuration
func (m *LocalSessionManager) Init(ctx context.Context, cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cfg = cfg
	return nil
}

// Start begins the session management background processes
func (m *LocalSessionManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("session manager already started")
	}

	m.started = true

	// Start background sync goroutine for running sessions
	go m.backgroundSync(ctx)

	// Initial pool setup: sync existing sessions and ensure minimum
	go func() {
		// First sync existing sessions from AMS
		if err := m.syncRunningSession(context.Background()); err != nil {
			logger.Errorf("failed to sync running sessions during startup: %v", err)
		}

		// Then ensure minimum pool size
		if err := m.ensureMinPoolSize(context.Background()); err != nil {
			logger.Errorf("failed to ensure min pool size during startup: %v", err)
		}
	}()

	return nil
}

// Stop stops the session manager
func (m *LocalSessionManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.started = false
	close(m.syncStopCh)

	return nil
}

// AcquireCold gets a cold session and changes status cold -> warming
func (m *LocalSessionManager) AcquireCold(ctx context.Context) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find a cold session
	for _, session := range m.cache {
		if session.Status == Cold {
			// Change status to warming
			session.Status = Warming
			session.LastHeartbeat = time.Now()
			return session, nil
		}
	}

	return nil, fmt.Errorf("no cold sessions available")
}

// SetWarmed changes session status from warming -> warmed
func (m *LocalSessionManager) SetWarmed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find session and check if it's warming
	session, exists := m.cache[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	if session.Status != Warming {
		return fmt.Errorf("session %s is not in warming status, current status: %s", id, session.Status)
	}

	// Change status to warmed
	session.Status = Warmed
	session.LastHeartbeat = time.Now()

	return nil
}

// AcquireWarmed gets a warmed session and changes status warmed -> in_use
func (m *LocalSessionManager) AcquireWarmed(ctx context.Context) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find a warmed session
	for _, session := range m.cache {
		if session.Status == Warmed {
			// Change status to in_use
			session.Status = InUse
			session.ExpiresAt = time.Now().Add(m.cfg.SessionTTL)
			session.LastHeartbeat = time.Now()
			return session, nil
		}
	}

	return nil, fmt.Errorf("no warmed sessions available")
}

// Release deletes a session completely
func (m *LocalSessionManager) Release(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.cache[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	// Remove from cache
	delete(m.cache, id)

	// Delete from anbox
	if session.Anbox != nil {
		// Use background context to avoid cancellation issues
		return m.anboxClient.Delete(context.Background(), session.Anbox.ID)
	}

	return nil
}

// GetSession retrieves a session by ID
func (m *LocalSessionManager) GetSession(ctx context.Context, id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.cache[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

// ListSessions returns all sessions with the specified status order by status
func (m *LocalSessionManager) ListSessions(ctx context.Context) ([]*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, session := range m.cache {
		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].Status == Cold {
			return true
		}
		if sessions[j].Status == Cold {
			return false
		}
		if sessions[i].Status == Warming {
			return true
		}
		if sessions[j].Status == Warming {
			return false
		}
		if sessions[i].Status == Warmed {
			return true
		}
		if sessions[j].Status == Warmed {
			return false
		}
		if sessions[i].Status == InUse {
			return true
		}
		if sessions[j].Status == InUse {
			return false
		}
		return false
	})

	return sessions, nil
}

// Heartbeat updates the last heartbeat time for a session
func (m *LocalSessionManager) Heartbeat(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.cache[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	session.LastHeartbeat = time.Now()
	return nil
}

// PoolStatus returns the current status of the session pool
func (m *LocalSessionManager) PoolStatus(ctx context.Context) (PoolStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := PoolStatus{Total: len(m.cache)}

	for _, session := range m.cache {
		switch session.Status {
		case Cold:
			status.Cold++
		case Warming:
			status.Warming++
		case Warmed:
			status.Warmed++
		case InUse:
			status.InUse++
		}
	}

	return status, nil
}

// syncRunningSession syncs running sessions from AMS
func (m *LocalSessionManager) syncRunningSession(ctx context.Context) error {
	runningSessionDetails, err := m.anboxClient.GetAllRunningSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get running sessions: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a map of running session IDs for quick lookup
	runningSessionMap := make(map[string]*anbox.SessionDetails)
	for _, session := range runningSessionDetails {
		runningSessionMap[session.ID] = session
	}

	// Add new running sessions that we don't have locally
	for sessionID, anboxSession := range runningSessionMap {
		if _, exists := m.cache[sessionID]; !exists {
			// Create new local session for running anbox session
			session := &Session{
				ID:            sessionID,
				Game:          m.cfg.GameName,
				GatewayURL:    m.anboxClient.GetGatewayURL(),
				AuthToken:     m.anboxClient.GetAuthToken(),
				Status:        Cold, // Start as cold, can be promoted later
				Anbox:         anboxSession,
				ExpiresAt:     time.Now().Add(m.cfg.SessionTTL),
				LastHeartbeat: time.Now(),
				CreatedAt:     time.Now(),
			}

			m.cache[sessionID] = session
		}
	}

	// Remove local sessions that are no longer running on AMS
	for sessionID := range m.cache {
		if _, exists := runningSessionMap[sessionID]; !exists {
			// Session is no longer running, remove it
			delete(m.cache, sessionID)
		}
	}

	return nil
}

// Helper methods

func (m *LocalSessionManager) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Check all sessions for expiration or heartbeat timeout
	for sessionID, session := range m.cache {
		shouldDelete := false

		// Check cold sessions for expiration
		if now.After(session.CreatedAt.Add(m.cfg.SessionTTL)) {
			shouldDelete = true
		}

		// Check in-use sessions for heartbeat timeout
		if session.Status == InUse || session.Status == Warmed {
			if now.Sub(session.LastHeartbeat) > m.cfg.HeartbeatTimeout {
				shouldDelete = true
			}
		}

		if shouldDelete {
			// Remove expired session and delete
			delete(m.cache, sessionID)
			logger.Warnf("session %s expired, deleting", sessionID)
			// Delete from anbox in background
			go func(s *Session) {
				if s.Anbox != nil {
					if err := m.anboxClient.Delete(context.Background(), s.Anbox.ID); err != nil {
						logger.Errorf("failed to delete anbox session %s: %v", s.Anbox.ID, err)
					}
				}
			}(session)
		}
	}
}

func (m *LocalSessionManager) backgroundSync(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.syncStopCh:
			return
		case <-ticker.C:
			// Sync running sessions from AMS
			if err := m.syncRunningSession(ctx); err != nil {
				logger.Errorf("failed to sync running sessions: %v", err)
			}

			// Cleanup expired sessions
			m.cleanupExpired()

			// Ensure minimum session pool size
			if err := m.ensureMinPoolSize(ctx); err != nil {
				logger.Errorf("failed to ensure min pool size: %v", err)
			}
		}
	}
}

// ensureMinPoolSize ensures the session pool has at least the minimum number of sessions
func (m *LocalSessionManager) ensureMinPoolSize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTotal := len(m.cache)

	// If we already have enough sessions, no need to create more
	if currentTotal >= m.cfg.Min {
		return nil
	}

	// Check if we've reached the maximum limit
	if currentTotal >= m.cfg.Max {
		logger.Warnf("session pool is at maximum capacity (%d), cannot create more sessions", m.cfg.Max)
		return nil
	}

	// 每次只创建一个否则,会批量一起过期
	go m.createNewSession(context.Background())

	return nil
}

// createNewSession creates a new session via anbox
func (m *LocalSessionManager) createNewSession(ctx context.Context) {
	req := anbox.CreateSessionRequest{
		App:      m.cfg.GameName,
		Joinable: true,
		Screen: anbox.Screen{
			Width:   m.cfg.ScreenConfig.Width,
			Height:  m.cfg.ScreenConfig.Height,
			Density: m.cfg.ScreenConfig.Density,
			FPS:     m.cfg.ScreenConfig.Fps,
		},
	}

	// Create session asynchronously via anbox
	if err := m.anboxClient.CreateAsync(ctx, req); err != nil {
		logger.Errorf("createNewSession failed to create session for game %s: %v", m.cfg.GameName, err)
		return
	}

	logger.Infof("createNewSession requested new session creation for game %s", m.cfg.GameName)
	// Note: The actual session will be picked up by the next sync cycle
}
