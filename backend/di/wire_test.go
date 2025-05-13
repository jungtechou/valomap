package di

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireRegistration(t *testing.T) {
	// Verify InjectorSet is defined (indirect testing of wire.go)
	assert.NotNil(t, InjectorSet, "InjectorSet should not be nil")
}

func TestBuildInjector(t *testing.T) {
	// Skip by default to avoid environment issues in CI
	if os.Getenv("TEST_BUILD_INJECTOR") != "1" {
		t.Skip("Skipping BuildInjector test. Set TEST_BUILD_INJECTOR=1 to enable")
	}

	// Build the injector
	injector, cleanup, err := BuildInjector()

	// Run cleanup at the end
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	// Check results
	require.NoError(t, err, "BuildInjector should not return an error")
	require.NotNil(t, injector, "Injector should not be nil")

	// Verify injector components
	assert.NotNil(t, injector.HttpEngine, "HttpEngine should not be nil")
	assert.NotNil(t, injector.Config, "Config should not be nil")
	assert.NotNil(t, injector.ImageCache, "ImageCache should not be nil")
	assert.NotNil(t, injector.HTTPClient, "HTTPClient should not be nil")
}

// Skipping TestBuildInjector because it requires file system permissions
// that aren't available in the test environment.
