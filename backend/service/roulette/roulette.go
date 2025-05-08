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
	ErrEmptyMapList = errors.New("received empty map list from API")
	ErrAPIRequest   = errors.New("failed to make API request")
	ErrAPIResponse  = errors.New("received invalid API response")
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

func (s *RouletteService) Roulette(ctx ctx.CTX) (*domain.Map, error) {
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

	// Select a random map using thread-safe RNG
	selectedMap := maps[s.rng.Intn(len(maps))]
	return &selectedMap, nil
}
