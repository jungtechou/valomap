package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvideConfig(t *testing.T) {
	// Test basic config functionality
	config, err := ProvideConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify config structure
	assert.NotEmpty(t, config.Server.Port)
	assert.NotZero(t, config.Server.ReadTimeout)
	assert.NotZero(t, config.Server.WriteTimeout)
	assert.NotZero(t, config.Server.ShutdownTimeout)

	assert.NotEmpty(t, config.Logging.Level)
	assert.NotEmpty(t, config.Logging.Format)

	assert.NotEmpty(t, config.API.BasePath)
	assert.NotEmpty(t, config.API.Version)
	assert.NotZero(t, config.API.RequestTimeout)
}

func TestInjectorStructFields(t *testing.T) {
	// Test that Injector struct has the expected fields
	injector := &Injector{}

	// Verify fields exist - reflection would be better but this simple check works too
	assert.NotPanics(t, func() {
		_ = injector.HttpEngine
		_ = injector.Config
		_ = injector.ImageCache
		_ = injector.HTTPClient
	})
}
