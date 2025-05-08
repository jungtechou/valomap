package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jungtechou/valomap/di"
	"github.com/sirupsen/logrus"
)

// Run initializes and starts the application
func Run() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Initialize dependency injection
	injector, cleanup, err := di.BuildInjector()
	if err != nil {
		log.WithError(err).Fatal("Failed to build injector")
	}
	defer cleanup()

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

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
