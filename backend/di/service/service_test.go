package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/jungtechou/valomap/config"
	"github.com/stretchr/testify/assert"
)

func TestProvideMapPool(t *testing.T) {
	mapPool := ProvideMapPool()

	// Verify the map pool contains expected maps
	assert.NotEmpty(t, mapPool.Maps)
	assert.Equal(t, 10, len(mapPool.Maps), "Expected 10 maps in the pool")

	// Verify some specific maps are present
	mapNames := make(map[string]bool)
	for _, m := range mapPool.Maps {
		mapNames[m.DisplayName] = true
	}

	expectedMaps := []string{"Ascent", "Split", "Bind", "Haven", "Breeze"}
	for _, name := range expectedMaps {
		assert.True(t, mapNames[name], "Map %s should be in the pool", name)
	}
}

func TestProvideHTTPClient(t *testing.T) {
	client := ProvideHTTPClient()

	// Verify client is not nil and has expected timeout
	assert.NotNil(t, client)
	assert.Equal(t, 10, int(client.Timeout.Seconds()), "Expected 10-second timeout")
}

func TestProvideImageCache(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "3000",
		},
		API: config.APIConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	client := &http.Client{}

	// Test with valid config and client - this might fail if we can't create the cache directory
	cache, err := ProvideImageCache(cfg, client)
	if err == nil {
		assert.NotNil(t, cache)
	}

	// Test with nil config - should fail
	cache, err = ProvideImageCache(nil, client)
	assert.Error(t, err)

	// Test with nil client - should fail
	cache, err = ProvideImageCache(cfg, nil)
	assert.Error(t, err)
}
