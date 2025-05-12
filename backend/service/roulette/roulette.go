package roulette

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service/cache"
	"github.com/sirupsen/logrus"
)

const (
	apiURL = "https://valorant-api.com/v1/maps"
)

var (
	ErrEmptyMapList   = errors.New("received empty map list from API")
	ErrAPIRequest     = errors.New("failed to make API request")
	ErrAPIResponse    = errors.New("received invalid API response")
	ErrNoStandardMaps = errors.New("no standard maps found")
	ErrNoFilteredMaps = errors.New("no maps found matching filter criteria")
)

type RouletteService struct {
	client     *http.Client
	rng        *rand.Rand // Thread-safe random number generator
	imageCache cache.ImageCache
}

func NewService(c *http.Client, imageCache cache.ImageCache) Service {
	// Create a new random source with current time seed
	source := rand.NewSource(time.Now().UnixNano())
	return &RouletteService{
		client:     c,
		rng:        rand.New(source),
		imageCache: imageCache,
	}
}

// fetchMaps fetches all maps from the API
func (s *RouletteService) fetchMaps(ctx ctx.CTX) ([]domain.Map, error) {
	// Make the API request
	resp, err := s.client.Get(apiURL)
	if err != nil {
		ctx.FieldLogger.WithFields(logrus.Fields{
			"error": err,
			"url":   apiURL,
		}).Error("Failed to fetch maps from API")
		return nil, fmt.Errorf("%w: %v", ErrAPIRequest, err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		ctx.FieldLogger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"url":         apiURL,
		}).Error("Received non-OK status code from API")
		return nil, fmt.Errorf("%w: status code %d", ErrAPIResponse, resp.StatusCode)
	}

	// Parse the response body
	var mapResp domain.MapResponse
	if err := json.NewDecoder(resp.Body).Decode(&mapResp); err != nil {
		ctx.FieldLogger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to decode API response")
		return nil, fmt.Errorf("%w: %v", ErrAPIResponse, err)
	}

	// Validate the response data
	maps := mapResp.Data
	if len(maps) == 0 {
		ctx.FieldLogger.Error("Received empty map list from API")
		return nil, ErrEmptyMapList
	}

	// Process map images via caching service if available
	if s.imageCache != nil {
		ctx.FieldLogger.Info("Processing map images through cache service")
		maps, err = s.imageCache.CacheMapImages(ctx, maps)
		if err != nil {
			ctx.FieldLogger.WithError(err).Warn("Error while caching map images, continuing with original URLs")
		}
	}

	// Log successful fetch
	ctx.FieldLogger.WithField("map_count", len(maps)).Info("Successfully fetched maps from API")

	return maps, nil
}

// filterMaps applies the provided filters to the map list
func (s *RouletteService) filterMaps(ctx ctx.CTX, maps []domain.Map, filter MapFilter) ([]domain.Map, error) {
	var filteredMaps []domain.Map

	// Create a set of banned map IDs for faster lookup
	bannedMaps := make(map[string]bool)
	for _, id := range filter.BannedMapIDs {
		bannedMaps[id] = true
	}

	// First filter out banned maps
	for _, m := range maps {
		if !bannedMaps[m.UUID] {
			filteredMaps = append(filteredMaps, m)
		}
	}

	// No maps left after removing banned maps
	if len(filteredMaps) == 0 {
		ctx.FieldLogger.WithFields(logrus.Fields{
			"filter": "banned maps",
			"banned_count": len(filter.BannedMapIDs),
		}).Error("All maps have been banned")
		return nil, ErrNoFilteredMaps
	}

	// Then apply standard filter if needed
	if filter.StandardOnly {
		var standardMaps []domain.Map

		for _, m := range filteredMaps {
			if m.TacticalDescription != "" {
				standardMaps = append(standardMaps, m)
			}
		}

		if len(standardMaps) == 0 {
			ctx.FieldLogger.WithField("filter", "standard").Error("No maps found matching filter criteria")
			return nil, ErrNoStandardMaps
		}

		ctx.FieldLogger.WithFields(logrus.Fields{
			"filter":         "standard",
			"filtered_count": len(standardMaps),
			"total_count":    len(filteredMaps),
		}).Info("Applied standard map filter")

		return standardMaps, nil
	}

	ctx.FieldLogger.WithFields(logrus.Fields{
		"banned_maps_count": len(filter.BannedMapIDs),
		"remaining_maps":   len(filteredMaps),
	}).Info("Applied map filters")

	return filteredMaps, nil
}

// GetRandomMap returns a random map filtered by the provided options
func (s *RouletteService) GetRandomMap(ctx ctx.CTX, filter MapFilter) (*domain.Map, error) {
	// Log filter options
	ctx.FieldLogger.WithFields(logrus.Fields{
		"standard_only": filter.StandardOnly,
	}).Debug("Map filter options")

	// Fetch all maps
	maps, err := s.fetchMaps(ctx)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filteredMaps, err := s.filterMaps(ctx, maps, filter)
	if err != nil {
		return nil, err
	}

	// Select a random map from the filtered list
	selectedMap := filteredMaps[s.rng.Intn(len(filteredMaps))]
	return &selectedMap, nil
}

// GetAllMaps returns all available maps
func (s *RouletteService) GetAllMaps(ctx ctx.CTX) ([]domain.Map, error) {
	// Log the request
	ctx.FieldLogger.Info("Fetching all maps")

	// Fetch all maps from the API
	maps, err := s.fetchMaps(ctx)
	if err != nil {
		return nil, err
	}

	// Log the result
	ctx.FieldLogger.WithField("map_count", len(maps)).Info("Successfully fetched all maps")

	return maps, nil
}
