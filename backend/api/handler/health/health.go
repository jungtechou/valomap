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
	Status    string    `json:"status" example:"ok"`
	Version   string    `json:"version" example:"1.0.0"`
	Uptime    string    `json:"uptime" example:"2h3m4s"`
	Timestamp time.Time `json:"timestamp" example:"2023-07-01T12:34:56Z"`
	GoVersion string    `json:"go_version" example:"go1.24"`
	Memory    Memory    `json:"memory"`
}

// Memory represents memory statistics
type Memory struct {
	Alloc      uint64 `json:"alloc" example:"1024"`
	TotalAlloc uint64 `json:"total_alloc" example:"2048"`
	Sys        uint64 `json:"sys" example:"4096"`
	NumGC      uint32 `json:"num_gc" example:"10"`
}

// NewHandler creates a new health check handler
func NewHandler() Handler {
	return &HealthHandler{}
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// HealthCheck godoc
// @Summary Health Check
// @Description Get the API health status with detailed system information
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse "Successful health check response"
// @Failure 500 {object} map[string]interface{} "Unexpected server error"
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
// @Description Get a simple pong response to check if the API is responsive
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Successful ping response with 'message': 'pong'"
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
