package api

import (
	"context"
	"time"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/letusgogo/playable-backend/internal/game"
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
	ctx         context.Context
	cancel      context.CancelFunc
	gameManager *game.Manager
}

func NewApiService(config ApiServiceConfig, gameManager *game.Manager) *ApiService {
	return &ApiService{
		name:        "apiService",
		config:      config,
		ginServer:   utils.NewGinServer(config.Address),
		gameManager: gameManager,
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
	v1 := a.ginServer.GinGroup("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		logger.GetLogger("apiService").Info("health check")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	gameGroup := v1.Group("/games")
	{
		gameGroup.GET("/:game", a.getGameInstance)
		gameGroup.GET("/:game/sessions", a.getGameInstanceSessions)

		// Session management endpoints - simplified
		gameGroup.POST("/:game/acquire_cold", a.acquireColdSession)
		gameGroup.POST("/:game/set_warmed", a.setSessionWarmed)
		gameGroup.POST("/:game/acquire_warmed", a.acquireWarmedSession)
		gameGroup.POST("/:game/release", a.releaseSession)

		gameGroup.POST("/:game/detect", a.detectStage)
	}
}

func (a *ApiService) detectStage(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	var req DetectStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CommonResponse{
			Code:    400,
			Message: "invalid request body",
			Data:    nil,
		})
		return
	}

	stageDetector := gameInstance.GetStageDetector(req.CurrentStageNum)
	match, evidence, err := stageDetector.Detect(c.Request.Context(), game, req.CurrentStageNum, req.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	response := DetectStageResponse{
		Match:    match,
		StageNum: req.CurrentStageNum,
		Evidence: evidence,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    response,
	})
}

func (a *ApiService) getGameInstance(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"error": "game not found"})
		return
	}
	status, err := gameInstance.GetInstanceStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    status,
	})
}

func (a *ApiService) getGameInstanceSessions(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	// Get pool status instead of listing sessions
	poolStatus, err := gameInstance.GetSessionManager().PoolStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    poolStatus,
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

// acquireColdSession 获取 cold session
func (a *ApiService) acquireColdSession(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	session, err := gameInstance.GetSessionManager().AcquireCold(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    session,
	})
}

// setSessionWarmed 设置 session 为 warmed 状态
func (a *ApiService) setSessionWarmed(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	var req SetWarmedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CommonResponse{
			Code:    400,
			Message: "invalid request body",
			Data:    nil,
		})
		return
	}

	err := gameInstance.GetSessionManager().SetWarmed(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    nil,
	})
}

// acquireWarmedSession 获取 warmed session
func (a *ApiService) acquireWarmedSession(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	session, err := gameInstance.GetSessionManager().AcquireWarmed(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    session,
	})
}

// releaseSession 删除 session
func (a *ApiService) releaseSession(c *gin.Context) {
	game := c.Param("game")
	gameInstance, ok := a.gameManager.GetGameInstance(c.Request.Context(), game)
	if !ok {
		c.JSON(http.StatusNotFound, CommonResponse{
			Code:    404,
			Message: "game not found",
			Data:    nil,
		})
		return
	}

	var req ReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CommonResponse{
			Code:    400,
			Message: "invalid request body",
			Data:    nil,
		})
		return
	}

	err := gameInstance.GetSessionManager().Release(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CommonResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    ErrNot,
		Message: "success",
		Data:    nil,
	})
}
