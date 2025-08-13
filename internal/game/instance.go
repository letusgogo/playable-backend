package game

import (
	"context"
	"fmt"
	"time"

	"github.com/letusgogo/playable-backend/internal/detector"
	"github.com/letusgogo/playable-backend/internal/session"
)

type GameInstance struct {
	gameConfig     *Game
	name           string
	anboxClient    session.AnboxClient
	sessionManager session.Manager
	initialized    bool
	running        bool
}

// NewGameInstance creates a new game instance with the given configuration
func NewGameInstance(gameConfig *Game, anboxClient session.AnboxClient) *GameInstance {
	return &GameInstance{
		gameConfig:  gameConfig,
		name:        gameConfig.Name,
		anboxClient: anboxClient,
		initialized: false,
		running:     false,
	}
}

// Init initializes the game instance's session manager
func (g *GameInstance) Init(ctx context.Context) error {
	if g.initialized {
		return fmt.Errorf("game instance %s already initialized", g.name)
	}
	if g.gameConfig.SessionConfig == nil {
		return fmt.Errorf("session config is nil")
	}

	// Convert game session config to session manager config
	sessionConfig := session.NewConfig()
	sessionConfig.GameName = g.gameConfig.Name
	sessionConfig.Min = g.gameConfig.SessionConfig.Min
	sessionConfig.Max = g.gameConfig.SessionConfig.Max
	sessionConfig.SessionTTL = g.gameConfig.Runtime.TimeOver
	sessionConfig.HeartbeatTimeout = 5 * time.Minute
	sessionConfig.SyncInterval = 30 * time.Second
	sessionConfig.ScreenConfig = &session.ScreenConfig{
		Width:   g.gameConfig.SessionConfig.ScreenConfig.Width,
		Height:  g.gameConfig.SessionConfig.ScreenConfig.Height,
		Density: g.gameConfig.SessionConfig.ScreenConfig.Density,
		Fps:     g.gameConfig.SessionConfig.ScreenConfig.Fps,
	}

	// Create session manager
	g.sessionManager = session.NewLocalSessionManager(sessionConfig, g.anboxClient)

	// Initialize session manager
	if err := g.sessionManager.Init(ctx, sessionConfig); err != nil {
		return fmt.Errorf("failed to initialize session manager for game %s: %w", g.name, err)
	}

	g.initialized = true
	return nil
}

// Start starts the game instance's session manager
func (g *GameInstance) Start(ctx context.Context) error {
	if !g.initialized {
		return fmt.Errorf("game instance %s not initialized", g.name)
	}

	if g.running {
		return fmt.Errorf("game instance %s already running", g.name)
	}

	if err := g.sessionManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start session manager for game %s: %w", g.name, err)
	}

	g.running = true
	return nil
}

// Stop stops the game instance's session manager
func (g *GameInstance) Stop(ctx context.Context) error {
	if !g.running {
		return nil
	}

	if err := g.sessionManager.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop session manager for game %s: %w", g.name, err)
	}

	g.running = false
	return nil
}

// GetSessionManager returns the session manager for this game instance
func (g *GameInstance) GetSessionManager() session.Manager {
	return g.sessionManager
}

// GetConfig returns the game configuration
func (g *GameInstance) GetConfig() *Game {
	return g.gameConfig
}

// IsInitialized returns whether the game instance is initialized
func (g *GameInstance) IsInitialized() bool {
	return g.initialized
}

// IsRunning returns whether the game instance is running
func (g *GameInstance) IsRunning() bool {
	return g.running
}

func (g *GameInstance) GetInstanceStatus(ctx context.Context) (*GameInstanceStatus, error) {
	poolStatus, err := g.sessionManager.PoolStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &GameInstanceStatus{
		Name:        g.name,
		Initialized: g.initialized,
		Running:     g.running,
		PoolStatus:  &poolStatus,
		Config:      g.gameConfig,
	}, nil
}

func (g *GameInstance) GetStageDetector(stageNum int) detector.StageChecker {
	return detector.NewDefaultOcrDetector()
}
