package game

import (
	"context"
	"fmt"
	"sync"

	"github.com/letusgogo/playable-backend/internal/session"
)

type Manager struct {
	gameInstances map[string]*GameInstance
	mu            sync.RWMutex
	anboxClient   session.AnboxClient
	initialized   bool
	running       bool
}

func NewManager(gameConfigs []*Game, anboxClient session.AnboxClient) *Manager {
	gameInstances := make(map[string]*GameInstance)
	for _, g := range gameConfigs {
		gameInstances[g.Name] = NewGameInstance(g, anboxClient)
	}
	return &Manager{
		gameInstances: gameInstances,
		anboxClient:   anboxClient,
		initialized:   false,
		running:       false,
	}
}

// Init initializes all game instances
func (m *Manager) Init(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("game manager already initialized")
	}

	// Initialize all game instances
	for gameName, instance := range m.gameInstances {
		if err := instance.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize game instance %s: %w", gameName, err)
		}
	}

	m.initialized = true
	return nil
}

// Start starts all game instances
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("game manager not initialized")
	}

	if m.running {
		return fmt.Errorf("game manager already running")
	}

	// Start all game instances
	for gameName, instance := range m.gameInstances {
		if err := instance.Start(ctx); err != nil {
			// If one instance fails to start, stop all already started instances
			m.stopAllInstances(ctx)
			return fmt.Errorf("failed to start game instance %s: %w", gameName, err)
		}
	}

	m.running = true
	return nil
}

// Stop stops all game instances
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.stopAllInstances(ctx)
	m.running = false
	return nil
}

// stopAllInstances stops all instances (internal helper method)
func (m *Manager) stopAllInstances(ctx context.Context) {
	for _, instance := range m.gameInstances {
		if err := instance.Stop(ctx); err != nil {
			// Log error but continue stopping other instances
			// Note: In a real implementation, you might want to use a proper logger
			fmt.Printf("error stopping game instance %s: %v\n", instance.name, err)
		}
	}
}

func (m *Manager) GetGameInstance(ctx context.Context, game string) (*GameInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.gameInstances[game]
	return g, ok
}

func (m *Manager) GetAllConfigs(ctx context.Context) []*Game {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]*Game, 0)
	for _, g := range m.gameInstances {
		games = append(games, g.gameConfig)
	}
	return games
}

// GetAllGameInstances returns all game instances
func (m *Manager) GetAllGameInstances(ctx context.Context) map[string]*GameInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	instances := make(map[string]*GameInstance)
	for name, instance := range m.gameInstances {
		instances[name] = instance
	}
	return instances
}

// GetAllGameInstancesStatus returns status of all game instances
func (m *Manager) GetAllGameInstancesStatus(ctx context.Context) (map[string]GameInstanceStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]GameInstanceStatus)

	for gameName, instance := range m.gameInstances {
		status := GameInstanceStatus{
			Name:        instance.name,
			Initialized: instance.IsInitialized(),
			Running:     instance.IsRunning(),
		}

		// Get pool status if instance is initialized
		if instance.IsInitialized() {
			poolStatus, err := instance.GetSessionManager().PoolStatus(ctx)
			if err != nil {
				// Continue with other instances even if one fails
				fmt.Printf("failed to get pool status for game %s: %v\n", gameName, err)
				statuses[gameName] = status
				continue
			}
			status.PoolStatus = &poolStatus
		}

		statuses[gameName] = status
	}

	return statuses, nil
}

// IsInitialized returns whether the manager is initialized
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// IsRunning returns whether the manager is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
