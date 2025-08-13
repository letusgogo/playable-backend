package anbox

type GatewayConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

// Screen represents the display configuration for a session
type Screen struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	Density int `json:"density"`
	FPS     int `json:"fps"`
}

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	App         string `json:"app"`
	AppVersion  int    `json:"app_version"`
	Ephemeral   bool   `json:"ephemeral"`
	ExtraData   string `json:"extra_data"`
	IdleTimeMin int    `json:"idle_time_min"`
	Joinable    bool   `json:"joinable"`
	Screen      Screen `json:"screen"`
}

// CreateSessionResponse represents the API response when creating a new session
type CreateSessionResponse struct {
	Type       string         `json:"type"`
	Status     string         `json:"status"`
	StatusCode int            `json:"status_code"`
	Metadata   SessionDetails `json:"metadata"`
}

// SessionDetails represents the session information returned by the API
type SessionDetails struct {
	ID          string       `json:"id"`
	Region      string       `json:"region"`
	URL         string       `json:"url"`
	StunServers []StunServer `json:"stun_servers"`
	Status      string       `json:"status"`
	Joinable    bool         `json:"joinable"`
}

// StunServer represents a STUN/TURN server configuration
type StunServer struct {
	URLs     []string `json:"urls"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}
