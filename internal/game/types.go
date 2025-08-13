package game

import (
	"time"

	"github.com/letusgogo/playable-backend/internal/detector"
	"github.com/letusgogo/playable-backend/internal/session"
)

type Config struct {
	Server Server       `mapstructure:"server"`
	Anbox  Anbox        `mapstructure:"anbox"`
	Games  []GameConfig `mapstructure:"games"`
}

type Server struct {
	Address string `mapstructure:"address"`
	Debug   bool   `mapstructure:"debug"`
}

type Anbox struct {
	Address    string `mapstructure:"address"`
	Token      string `mapstructure:"token"`
	AMSAddress string `mapstructure:"ams_address"`
}

type GameConfig struct {
	Name          string            `mapstructure:"name"`
	SessionConfig *SessionConfig    `mapstructure:"session_config"`
	Runtime       *Runtime          `mapstructure:"runtime"`
	Stages        []*detector.Stage `mapstructure:"stages"`
}

type SessionConfig struct {
	Min              int           `mapstructure:"min"`
	Max              int           `mapstructure:"max"`
	SessionTTL       time.Duration `mapstructure:"session_ttl"`
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat_timeout"`
	SyncInterval     time.Duration `mapstructure:"sync_interval"`
	ScreenConfig     ScreenConfig  `mapstructure:"screen_config"`
}

type ScreenConfig struct {
	Width   int `mapstructure:"width"`
	Height  int `mapstructure:"height"`
	Density int `mapstructure:"density"`
	Fps     int `mapstructure:"fps"`
}

type Runtime struct {
	TimeOver time.Duration `mapstructure:"time_over"`
	OverURL  string        `mapstructure:"over_url"`
}

// GameInstanceStatus represents the status of a game instance
type GameInstanceStatus struct {
	Name        string              `json:"name"`
	Initialized bool                `json:"initialized"`
	Running     bool                `json:"running"`
	PoolStatus  *session.PoolStatus `json:"pool_status,omitempty"`
	Config      *GameConfig         `json:"config,omitempty"`
}
