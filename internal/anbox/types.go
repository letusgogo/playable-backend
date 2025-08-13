package anbox

type AnboxConfig struct {
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
	AmsAddr string `mapstructure:"ams_address"`
	AmsCert string `mapstructure:"ams_cert"`
	AmsKey  string `mapstructure:"ams_key"`
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

// ListInstanceDetails contains only the essential instance information
type ListInstanceDetails struct {
	InstanceIDs []string
	TotalCount  int
}

// InstanceConfig represents the configuration of an instance
type InstanceConfig struct {
	Platform string `json:"platform"`
	Display  struct {
		Width   int `json:"width"`
		Height  int `json:"height"`
		FPS     int `json:"fps"`
		Density int `json:"density"`
	} `json:"display"`
}

// InstanceResources represents resource allocation for an instance
type InstanceResources struct {
	CPUs     int   `json:"cpus"`
	Memory   int64 `json:"memory"`
	DiskSize int64 `json:"disk-size"`
	GPUSlots int   `json:"gpu-slots"`
	VPUSlots int   `json:"vpu-slots"`
}

// InstanceDetails represents detailed information about an Anbox instance
type InstanceDetails struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Base          bool              `json:"base"`
	Type          string            `json:"type"`
	StatusCode    int               `json:"status_code"`
	Status        string            `json:"status"`
	Node          string            `json:"node"`
	AppID         string            `json:"app_id"`
	AppName       string            `json:"app_name"`
	AppVersion    int               `json:"app_version"`
	CreatedAt     int64             `json:"created_at"`
	Address       string            `json:"address"`
	PublicAddress string            `json:"public_address"`
	ErrorMessage  string            `json:"error_message"`
	StatusMessage string            `json:"status_message"`
	Config        InstanceConfig    `json:"config"`
	Resources     InstanceResources `json:"resources"`
	Architecture  string            `json:"architecture"`
	Tags          []string          `json:"tags"`
}

// ListInstancesResponse represents the response from AMS list instances API
type ListInstancesResponse struct {
	Type       string   `json:"type"`
	TotalSize  int      `json:"total_size"`
	Status     string   `json:"status"`
	StatusCode int      `json:"status_code"`
	ErrorCode  int      `json:"error_code"`
	Metadata   []string `json:"metadata"`
}

// InstanceDetailsResponse represents the response from AMS get instance details API
type InstanceDetailsResponse struct {
	Type       string          `json:"type"`
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	ErrorCode  int             `json:"error_code"`
	Metadata   InstanceDetails `json:"metadata"`
}
