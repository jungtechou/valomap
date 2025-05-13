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
			AllowedOrigins: []string{"http://localhost:3000"},
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

func TestInitialize(t *testing.T) {
	// Test case 1: Default mode
	t.Run("Default mode", func(t *testing.T) {
		// Create test config
		cfg := &config.Config{
			Logging: config.LoggingConfig{
				Level: "info",
			},
			Security: config.SecurityConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
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

		// Create engine without initialization
		engine := &GinEngine{
			config: cfg,
		}

		// Initialize engine
		engine.Initialize(testRouter)

		// Assertions
		assert.NotNil(t, engine.engine)
		assert.Equal(t, gin.ReleaseMode, gin.Mode())
	})

	// Test case 2: Debug mode
	t.Run("Debug mode", func(t *testing.T) {
		// Create test config
		cfg := &config.Config{
			Logging: config.LoggingConfig{
				Level: "debug",
			},
			Security: config.SecurityConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
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

		// Create engine without initialization
		engine := &GinEngine{
			config: cfg,
		}

		// Initialize engine
		engine.Initialize(testRouter)

		// Assertions
		assert.NotNil(t, engine.engine)
		assert.Equal(t, gin.DebugMode, gin.Mode())
	})

	// Test case 3: No config (default settings)
	t.Run("No config", func(t *testing.T) {
		// Create test router
		testRouter := &mockRouter{}

		// Create engine without config
		engine := &GinEngine{
			config: nil,
		}

		// Initialize engine
		engine.Initialize(testRouter)

		// Assertions
		assert.NotNil(t, engine.engine)
	})
}

func TestServeHTTP(t *testing.T) {
	// Setup Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test config with proper CORS settings
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Security: config.SecurityConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
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
		Security: config.SecurityConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		},
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

func TestStartServer(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "0", // Use a random free port
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		},
		Security: config.SecurityConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		},
	}

	// Create a test router
	testRouter := &mockRouter{}

	// Create engine
	engine := NewEngine(testRouter, cfg)

	// Create a goroutine to start the server
	errChan := make(chan error, 1)
	go func() {
		// Start the server - this will block until it's shut down
		err := engine.StartServer()
		errChan <- err
	}()

	// Wait a moment for the server to start
	time.Sleep(100 * time.Millisecond)

	// Shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := engine.GracefulShutdown(ctx)
	assert.NoError(t, err, "Shutdown should succeed")

	// Wait for the server to exit and check if there was an error
	select {
	case err := <-errChan:
		// Either we get ErrServerClosed or nil, both are acceptable
		if err != nil {
			assert.Equal(t, http.ErrServerClosed, err, "Server should close gracefully")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not shut down within the expected time")
	}
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
