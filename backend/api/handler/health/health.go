package health

import (
	"net/http"
	"runtime"
	"time"

	"github.com/jungtechou/valomap/api/handler"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
	Timestamp time.Time `json:"timestamp"`
	GoVersion string    `json:"go_version"`
	Memory    Memory    `json:"memory"`
}

// Memory represents memory statistics
type Memory struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// NewHandler creates a new health check handler
func NewHandler() Handler {
	return &HealthHandler{}
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// HealthCheck godoc
// @Summary Health Check
// @Description Get the API health status
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Create health response
	health := HealthResponse{
		Status:    "ok",
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Timestamp: time.Now(),
		GoVersion: runtime.Version(),
		Memory: Memory{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
	}

	c.JSON(http.StatusOK, health)
}

// Ping godoc
// @Summary Simple ping endpoint
// @Description Get a simple pong response
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// GetRouteInfos implements handler.Handler interface
func (h *HealthHandler) GetRouteInfos() []handler.RouteInfo {
	return []handler.RouteInfo{
		{
			Method:      http.MethodGet,
			Path:        "/health",
			Middlewares: []gin.HandlerFunc{},
			Handler:     h.HealthCheck,
		},
		{
			Method:      http.MethodGet,
			Path:        "/ping",
			Middlewares: []gin.HandlerFunc{},
			Handler:     h.Ping,
		},
	}
}
