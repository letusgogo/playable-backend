package session

import (
	"context"
	"fmt"
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
	"github.com/letusgogo/playable-backend/internal/game"
)

type Config struct {
}

type SessionManager interface {
	Create(ctx context.Context, gameName string) (*Session, error)
	Delete(ctx context.Context, gameName, id string) error
	GetAndSet(ctx context.Context, gameName string, status SessionStatus) (*Session, error)
}

type CacheSessionManager struct {
	cache              map[string]*Session
	gameManager        *game.Manager
	anboxGatewayClient *anbox.GatewayClient
}

func NewCacheSessionManager(gameManager *game.Manager, anboxGatewayClient *anbox.GatewayClient) *CacheSessionManager {
	return &CacheSessionManager{
		cache:              make(map[string]*Session),
		gameManager:        gameManager,
		anboxGatewayClient: anboxGatewayClient,
	}
}

func (m *CacheSessionManager) Create(ctx context.Context, gameName string) (*Session, error) {
	// get the game
	g, ok := m.gameManager.Get(ctx, gameName)
	if !ok {
		return nil, fmt.Errorf("game not found")
	}

	session, err := m.anboxGatewayClient.Create(ctx, anbox.CreateSessionRequest{
		App:         g.Name,
		Ephemeral:   true,
		IdleTimeMin: 3,
		Joinable:    true,
		Screen: anbox.Screen{
			Width:   g.SessionConfig.ScreenConfig.Width,
			Height:  g.SessionConfig.ScreenConfig.Height,
			Density: g.SessionConfig.ScreenConfig.Density,
			FPS:     g.SessionConfig.ScreenConfig.Fps,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:            session.ID,
		Status:        SessionStatus(session.Status),
		Anbox:         session,
		ExpiresAt:     time.Now().Add(time.Minute * 10),
		LastHeartbeat: time.Now(),
		CreatedAt:     time.Now(),
	}, nil
}

func (m *CacheSessionManager) Delete(ctx context.Context, game, id string) error {
	return nil
}

func (m *CacheSessionManager) GetAndSet(ctx context.Context, game string, status SessionStatus) (*Session, error) {
	// create a new session
	session, err := m.Create(ctx, game)
	if err != nil {
		return nil, err
	}

	// set the session status
	return session, nil
}
