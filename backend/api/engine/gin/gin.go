package gin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jungtechou/valomap/api/middleware"
	"github.com/jungtechou/valomap/api/router"
	"github.com/jungtechou/valomap/config"
	"github.com/jungtechou/valomap/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinEngine implements the Engine interface using Gin framework
type GinEngine struct {
	engine *gin.Engine
	server *http.Server
	config *config.Config
}

// NewEngine creates a new Gin engine instance
func NewEngine(r router.Router, cfg *config.Config) *GinEngine {
	engine := &GinEngine{
		config: cfg,
	}
	engine.Initialize(r)
	return engine
}

// Initialize sets up the Gin engine with middleware and routes
func (g *GinEngine) Initialize(r router.Router) {
	// Set Gin mode based on config
	if g.config != nil && g.config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a new Gin engine
	engine := gin.New()

	// Add middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.ErrorHandler())
	engine.Use(middleware.RequestContext())

	// Add CORS middleware if config exists
	if g.config != nil {
		corsConfig := cors.Config{
			AllowOrigins:     g.config.Security.AllowedOrigins,
			AllowMethods:     g.config.Security.AllowedMethods,
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
		engine.Use(cors.New(corsConfig))
	} else {
		// Default CORS settings
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		engine.Use(cors.New(corsConfig))
	}

	// Set maximum multipart memory
	if g.config != nil {
		engine.MaxMultipartMemory = g.config.Security.MaxBodySize
	} else {
		engine.MaxMultipartMemory = 8 << 20 // 8 MB default
	}

	// Register API routes
	r.RegisterAPI(engine)

	// Register Swagger documentation
	docs.RegisterSwagger(engine)

	// Store engine
	g.engine = engine
}

// StartServer starts the HTTP server
func (g *GinEngine) StartServer() error {
	// Configure server
	port := "3000" // Default port
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second

	// Use config if available
	if g.config != nil {
		port = g.config.Server.Port
		readTimeout = g.config.Server.ReadTimeout
		writeTimeout = g.config.Server.WriteTimeout
	}

	// Create and configure server
	g.server = &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      g.engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// Log start message
	logrus.WithField("port", port).Info("Starting HTTP server")

	// Start server
	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GracefulShutdown gracefully shuts down the server
func (g *GinEngine) GracefulShutdown(ctx context.Context) error {
	logrus.Info("Shutting down server...")
	return g.server.Shutdown(ctx)
}

// ServeHTTP implements the http.Handler interface
func (g *GinEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	g.engine.ServeHTTP(w, req)
}
