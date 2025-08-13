package game

import (
	"time"

	"github.com/letusgogo/playable-backend/internal/session"
)

type Config struct {
	Server Server `mapstructure:"server"`
	Anbox  Anbox  `mapstructure:"anbox"`
	Games  []Game `mapstructure:"games"`
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

type Game struct {
	Name          string         `mapstructure:"name"`
	SessionConfig *SessionConfig `mapstructure:"session_config"`
	Runtime       *Runtime       `mapstructure:"runtime"`
	Stages        []*Stage       `mapstructure:"stages"`
}

type SessionConfig struct {
	Min          int          `mapstructure:"min"`
	Max          int          `mapstructure:"max"`
	ScreenConfig ScreenConfig `mapstructure:"screen_config"`
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

type Area struct {
	Clue   string  `mapstructure:"clue"`
	X      float64 `mapstructure:"x"`
	Y      float64 `mapstructure:"y"`
	Width  float64 `mapstructure:"width"`
	Height float64 `mapstructure:"height"`
}

type Reco struct {
	Method string   `mapstructure:"method"`
	Matchs []string `mapstructure:"matchs"`
}

type Stage struct {
	Number   int           `mapstructure:"number"`
	Interval time.Duration `mapstructure:"interval"`
	Area     Area          `mapstructure:"area"`
	Reco     Reco          `mapstructure:"reco"`
}

// GameInstanceStatus represents the status of a game instance
type GameInstanceStatus struct {
	Name        string              `json:"name"`
	Initialized bool                `json:"initialized"`
	Running     bool                `json:"running"`
	PoolStatus  *session.PoolStatus `json:"pool_status,omitempty"`
	Config      *Game               `json:"config,omitempty"`
}
