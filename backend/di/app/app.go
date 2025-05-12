package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jungtechou/valomap/di"
	"github.com/jungtechou/valomap/service/cache"
	"github.com/sirupsen/logrus"
)

// Run initializes and starts the application
func Run() {
	log := setupLogger()

	// Initialize dependency injection
	injector, cleanup, err := setupDependencies(log)
	if err != nil {
		log.WithError(err).Fatal("Failed to build injector")
		return // Exit for testability
	}

	defer func() {
		shutdownApp(log, injector, cleanup)
	}()

	// Prewarm cache if possible
	if shouldPrewarmCache(injector) {
		prewarmCache(log, injector)
	} else {
		logCacheSkipReason(log, injector)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server and handle signals
	runServer(ctx, log, injector)
}

// setupLogger creates and configures the application logger
func setupLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return log
}

// setupDependencies initializes the dependency injection container
func setupDependencies(log *logrus.Logger) (*di.Injector, func(), error) {
	injector, cleanup, err := di.BuildInjector()
	if err != nil {
		return nil, nil, err
	}

	// Log configuration if available
	if injector.Config != nil {
		log.WithFields(logrus.Fields{
			"server_port": injector.Config.Server.Port,
			"log_level":   injector.Config.Logging.Level,
			"environment": func() string {
				if gin := os.Getenv("GIN_MODE"); gin == "release" {
					return "production"
				}
				return "development"
			}(),
		}).Info("Application configured")
	}

	return injector, cleanup, nil
}

// shouldPrewarmCache determines if caching prewarming should be attempted
func shouldPrewarmCache(injector *di.Injector) bool {
	return injector != nil &&
		injector.Config != nil &&
		injector.ImageCache != nil &&
		injector.HTTPClient != nil &&
		injector.Config.API.MapAPIURL != ""
}

// prewarmCache initializes the cache with map images
func prewarmCache(log *logrus.Logger, injector *di.Injector) {
	log.Info("Prewarming map image cache")
	prewarmer := cache.NewMapPrewarmer(injector.ImageCache, injector.HTTPClient, injector.Config.API.MapAPIURL)

	go func() {
		if err := prewarmer.PrewarmMapImages(); err != nil {
			log.WithError(err).Warn("Cache prewarming encountered an error")
		} else {
			log.Info("Cache prewarming completed successfully")
		}
	}()
}

// logCacheSkipReason logs why cache prewarming was skipped
func logCacheSkipReason(log *logrus.Logger, injector *di.Injector) {
	log.Warn("Cache prewarming skipped due to missing dependencies")
	if injector.ImageCache == nil {
		log.Warn("ImageCache is nil")
	}
	if injector.HTTPClient == nil {
		log.Warn("HTTPClient is nil")
	}
	if injector.Config == nil || injector.Config.API.MapAPIURL == "" {
		log.Warn("MapAPIURL is empty or Config is nil")
	}
}

// shutdownApp cleans up resources when the application exits
func shutdownApp(log *logrus.Logger, injector *di.Injector, cleanup func()) {
	// Shutdown image cache if available
	if injector != nil && injector.ImageCache != nil {
		log.Info("Shutting down image cache")
		injector.ImageCache.Shutdown()
	}
	// Run cleanup
	if cleanup != nil {
		cleanup()
	}
}

// runServer starts the HTTP server and handles shutdown signals
func runServer(ctx context.Context, log *logrus.Logger, injector *di.Injector) {
	// Handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server")
		if err := injector.HttpEngine.StartServer(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		log.WithError(err).Error("Server error")
	case sig := <-signalChan:
		log.WithField("signal", sig.String()).Info("Received shutdown signal")

		// Create a timeout context for graceful shutdown
		shutdownTimeout := 5 * time.Second
		if injector.Config != nil {
			shutdownTimeout = injector.Config.Server.ShutdownTimeout
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
		defer shutdownCancel()

		if err := injector.HttpEngine.GracefulShutdown(shutdownCtx); err != nil {
			log.WithError(err).Error("Error during shutdown")
		}
	}

	log.Info("Server stopped")
}
