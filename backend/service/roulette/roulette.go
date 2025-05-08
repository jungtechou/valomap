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
	client *http.Client
	rng    *rand.Rand // Thread-safe random number generator
}

func NewService(c *http.Client) Service {
	// Create a new random source with current time seed
	source := rand.NewSource(time.Now().UnixNano())
	return &RouletteService{
		client: c,
		rng:    rand.New(source),
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

	// Log successful fetch
	ctx.FieldLogger.WithField("map_count", len(maps)).Info("Successfully fetched maps from API")

	return maps, nil
}

// filterMaps applies the provided filters to the map list
func (s *RouletteService) filterMaps(ctx ctx.CTX, maps []domain.Map, filter MapFilter) ([]domain.Map, error) {
	if !filter.StandardOnly {
		return maps, nil
	}

	var filteredMaps []domain.Map
	for _, m := range maps {
		if m.TacticalDescription != "" {
			filteredMaps = append(filteredMaps, m)
		}
	}

	if len(filteredMaps) == 0 {
		ctx.FieldLogger.WithField("filter", "standard").Error("No maps found matching filter criteria")
		return nil, ErrNoStandardMaps
	}

	ctx.FieldLogger.WithFields(logrus.Fields{
		"filter":         "standard",
		"filtered_count": len(filteredMaps),
		"total_count":    len(maps),
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
