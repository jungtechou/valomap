package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/api/handler"
	"github.com/jungtechou/valomap/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Security: config.SecurityConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		},
		Server: config.ServerConfig{
			Port:         "3000",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// Create test router
	testRouter := &mockRouter{}

	// Test with valid engine options
	engine := NewEngine(testRouter, cfg)

	require.NotNil(t, engine)
	assert.IsType(t, &GinEngine{}, engine)
	assert.NotNil(t, engine.engine)
}

func TestServeHTTP(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
	}

	// Create test router
	testRouter := &mockRouter{}

	// Create engine
	engine := NewEngine(testRouter, cfg)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ping", nil)

	// Call ServeHTTP
	engine.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGracefulShutdown(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "0", // Use port 0 for testing to get a random available port
		},
	}

	// Create test router
	testRouter := &mockRouter{}

	// Create engine
	engine := NewEngine(testRouter, cfg)

	// Set up a test server
	engine.server = &http.Server{
		Addr:    ":0",
		Handler: engine,
	}

	// Call GracefulShutdown
	ctx := context.Background()
	err := engine.GracefulShutdown(ctx)

	// Assertions - should be no error on shutdown of a non-started server
	assert.NoError(t, err)
}

// Mock implementation of router.Router
type mockRouter struct {}

func (r *mockRouter) RegisterAPI(engine *gin.Engine) {
	// Add a simple ping handler for testing
	engine.GET("/api/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}

func (r *mockRouter) GetRoutesInfo() []handler.RouteInfo {
	return []handler.RouteInfo{
		{
			Method:      "GET",
			Path:        "/ping",
			Handler:     func(c *gin.Context) { c.String(http.StatusOK, "pong") },
			Middlewares: []gin.HandlerFunc{},
		},
	}
}

func (r *mockRouter) PrefixPath() string {
	return "api/v1"
}
