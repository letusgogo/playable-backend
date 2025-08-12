package internal

import (
	"context"
	"time"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/letusgogo/quick/logger"
	"github.com/letusgogo/quick/utils"
)

type ApiServiceConfig struct {
	Address string `yaml:"address"`
}

func NewApiServiceConfig() ApiServiceConfig {
	return ApiServiceConfig{
		Address: "0.0.0.0:2222",
	}
}

type ApiService struct {
	name      string
	config    ApiServiceConfig
	ginServer *utils.GinService
	// context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

func NewApiService(config ApiServiceConfig) *ApiService {
	return &ApiService{
		name:      "apiService",
		config:    config,
		ginServer: utils.NewGinServer(config.Address),
	}
}

func (a *ApiService) Name() string {
	return a.name
}

func (a *ApiService) Init() error {
	// Create context for graceful shutdown
	a.ctx, a.cancel = context.WithCancel(context.Background())

	// Setup API routes
	a.setupRoutes()

	return nil
}

func (a *ApiService) setupRoutes() {
	// Apply CORS middleware to the entire Gin engine
	a.ginServer.GinEngine().Use(cors.Default())

	a.ginServer.GinGroup("/api/v1").GET("/health", func(c *gin.Context) {
		logger.GetLogger("apiService").Info("health check")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
}

// Start starts the API service
func (a *ApiService) Start() error {

	go func() {
		if err := a.ginServer.Start(); err != nil {
			logger.GetLogger("apiService").Errorf("failed to start gin server: %v", err)
		}
	}()
	return nil
}

// StopGracefully stops the API service gracefully
func (a *ApiService) StopGracefully(wait time.Duration) error {
	logger.GetLogger("apiService").Info("stop api service")

	// Cancel context to signal shutdown
	if a.cancel != nil {
		a.cancel()
	}

	// Stop gin server
	return a.ginServer.Stop(wait)
}
