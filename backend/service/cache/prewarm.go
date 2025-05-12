package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
)

// MapPrewarmer handles prewarming of map images
type MapPrewarmer struct {
	cache  ImageCache
	client *http.Client
	apiURL string
}

// NewMapPrewarmer creates a new map prewarming service
func NewMapPrewarmer(cache ImageCache, client *http.Client, apiURL string) *MapPrewarmer {
	return &MapPrewarmer{
		cache:  cache,
		client: client,
		apiURL: apiURL,
	}
}

// PrewarmMapImages fetches all maps and caches their images
func (p *MapPrewarmer) PrewarmMapImages() error {
	logger := logrus.WithField("component", "MapPrewarmer")
	logger.Info("Starting map image prewarming")
	startTime := time.Now()

	// Create context
	reqCtx := ctx.Background()

	// Fetch maps from API
	resp, err := p.client.Get(p.apiURL)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch maps for prewarming")
		return err
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
		logger.WithError(err).Error("Failed to fetch maps for prewarming")
		return err
	}

	// Parse response
	var mapResp domain.MapResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&mapResp); err != nil {
		logger.WithError(err).Error("Failed to decode map response for prewarming")
		return err
	}

	maps := mapResp.Data
	logger.WithField("map_count", len(maps)).Info("Fetched maps for prewarming")

	// Extract image URLs to cache
	imageURLs := make(map[string]string)
	for _, m := range maps {
		if m.Splash != "" {
			imageURLs["map_"+m.UUID+"_splash"] = m.Splash
		}
		if m.DisplayIcon != "" {
			imageURLs["map_"+m.UUID+"_icon"] = m.DisplayIcon
		}
	}

	// Prewarm the cache
	if err := p.cache.PrewarmCache(reqCtx, imageURLs); err != nil {
		logger.WithError(err).Error("Error during cache prewarming")
		return err
	}

	duration := time.Since(startTime)
	logger.WithFields(logrus.Fields{
		"duration_ms": duration.Milliseconds(),
		"image_count": len(imageURLs),
	}).Info("Map image prewarming completed")

	return nil
}
