package session

import (
	"context"
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
)

// AnboxClient defines the interface for interacting with Anbox Gateway
// This allows for easier testing by providing a mockable interface
type AnboxClient interface {
	CreateAsync(ctx context.Context, req anbox.CreateSessionRequest) error
	Delete(ctx context.Context, sessionID string) error
	GetAllRunningSession(ctx context.Context) ([]*anbox.SessionDetails, error)
	GetGatewayURL() string
	GetAuthToken() string
}

type PoolStatus struct {
	Total   int `json:"total"`
	Cold    int `json:"cold"`
	Warming int `json:"warming"`
	Warmed  int `json:"warmed"`
	InUse   int `json:"in_use"`
}

type Config struct {
	GameName         string        `mapstructure:"game_name"`
	Min              int           `mapstructure:"min"`               // Minimum sessions to maintain
	Max              int           `mapstructure:"max"`               // Maximum total sessions allowed
	SessionTTL       time.Duration `mapstructure:"session_ttl"`       // Time before session expires
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat_timeout"` // Time before session considered dead
	SyncInterval     time.Duration `mapstructure:"sync_interval"`     // How often to sync running sessions from AMS
	ScreenConfig     *ScreenConfig `mapstructure:"screen_config"`
}

func NewConfig() *Config {
	return &Config{
		GameName:         "idle_weapon",
		Min:              5,
		Max:              10,
		SessionTTL:       5 * time.Minute,
		HeartbeatTimeout: 5 * time.Minute,
		SyncInterval:     30 * time.Second,
		ScreenConfig: &ScreenConfig{
			Width:   720,
			Height:  1240,
			Density: 320,
			Fps:     30,
		},
	}
}

type ScreenConfig struct {
	Width   int `mapstructure:"width"`
	Height  int `mapstructure:"height"`
	Density int `mapstructure:"density"`
	Fps     int `mapstructure:"fps"`
}

type SessionStatus string

const (
	Cold    SessionStatus = "cold"
	Warming SessionStatus = "warming"
	Warmed  SessionStatus = "warmed"
	InUse   SessionStatus = "in_use"
)

type Session struct {
	ID            string
	Game          string
	Status        SessionStatus
	Anbox         *anbox.SessionDetails
	GatewayURL    string
	AuthToken     string
	ExpiresAt     time.Time // InUse 的业务 TTL
	LastHeartbeat time.Time
	CreatedAt     time.Time
}
