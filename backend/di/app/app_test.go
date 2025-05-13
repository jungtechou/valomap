package app

import (
	"testing"

	"github.com/jungtechou/valomap/config"
	"github.com/jungtechou/valomap/di"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSetupLogger tests the setupLogger function
func TestSetupLogger(t *testing.T) {
	logger := setupLogger()
	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.Logger{}, logger)
	formatter, ok := logger.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok)
	assert.True(t, formatter.FullTimestamp)
}

// TestShouldPrewarmCache tests the shouldPrewarmCache function
func TestShouldPrewarmCache(t *testing.T) {
	// Test nil injector
	assert.False(t, shouldPrewarmCache(nil))

	// Test incomplete injector
	injector := &di.Injector{}
	assert.False(t, shouldPrewarmCache(injector))

	// Test with config but missing other components
	injector.Config = &config.Config{
		API: config.APIConfig{
			MapAPIURL: "https://example.com/api/maps",
		},
	}
	assert.False(t, shouldPrewarmCache(injector))
}

// TestLogCacheSkipReason tests the logCacheSkipReason function
func TestLogCacheSkipReason(t *testing.T) {
	logger := logrus.New()

	// Special handling for nil injector
	injector := &di.Injector{}
	logCacheSkipReason(logger, injector)

	// No assertion needed as we're just ensuring the function doesn't panic
}

// TestShutdownApp tests the shutdownApp function
func TestShutdownApp(t *testing.T) {
	logger := logrus.New()

	// Test with nil injector
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}
	shutdownApp(logger, nil, cleanup)
	assert.True(t, cleanupCalled)
}
